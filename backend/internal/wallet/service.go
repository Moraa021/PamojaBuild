package wallet

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"PamojaBuild/internal/bitcoin"
	"PamojaBuild/internal/config"
	"PamojaBuild/internal/lightning"
	"PamojaBuild/internal/models"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type Service struct {
	repo      Repository
	lndClient *lightning.Client
	config    *config.Config
	btcClient *bitcoin.Client
}

func NewService(repo Repository, lndClient *lightning.Client) *Service {
	cfg := config.Load()
	btcClient := bitcoin.NewClient(cfg.BTCRPCHost, cfg.BTCRPCUser, cfg.BTCRPCPass)
	return &Service{
		repo:      repo,
		lndClient: lndClient,
		config:    cfg,
		btcClient: btcClient,
	}
}

func (s *Service) CompleteTask(ctx context.Context, taskID int64, creatorUserID int64) (*models.PayoutRequest, error) {
	// 1. Verify caller is task poster
	taskCreator, taskStatus, err := s.repo.GetTaskCreatorAndStatus(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if taskCreator != creatorUserID {
		return nil, fmt.Errorf("user is not the creator of this task")
	}

	// 2. Verify task status is in_progress
	if taskStatus != "in_progress" {
		return nil, fmt.Errorf("task is not in_progress (status: %s)", taskStatus)
	}

	// 3. Calculate total_sats by summing all confirmed donations
	totalSats, err := s.repo.GetConfirmedDonationsTotal(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get confirmed donations total: %w", err)
	}
	if totalSats <= 0 {
		return nil, fmt.Errorf("cannot complete task with 0 confirmed donations")
	}

	// 4. Fetch all keyholder public keys
	keyholders, err := s.repo.GetKeyholders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyholders: %w", err)
	}
	if len(keyholders) != 5 {
		return nil, fmt.Errorf("expected 5 keyholders, got %d", len(keyholders))
	}

	// Parse and sort keys
	pubKeys := make([]*btcec.PublicKey, len(keyholders))
	for i, kh := range keyholders {
		b, err := hex.DecodeString(kh.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public key for keyholder %d: %w", kh.ID, err)
		}
		pk, err := btcec.ParsePubKey(b)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key for keyholder %d: %w", kh.ID, err)
		}
		pubKeys[i] = pk
	}

	sort.Slice(pubKeys, func(i, j int) bool {
		return bytes.Compare(pubKeys[i].SerializeCompressed(), pubKeys[j].SerializeCompressed()) < 0
	})

	// 5. Construct P2WSH 3-of-5 multisig address
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_3)
	for _, pk := range pubKeys {
		builder.AddData(pk.SerializeCompressed())
	}
	builder.AddOp(txscript.OP_5)
	builder.AddOp(txscript.OP_CHECKMULTISIG)
	multisigWitnessScript, err := builder.Script()
	if err != nil {
		return nil, fmt.Errorf("failed to build multisig script: %w", err)
	}

	witnessScriptHash := sha256.Sum256(multisigWitnessScript)
	p2wshAddr, err := btcutil.NewAddressWitnessScriptHash(witnessScriptHash[:], &chaincfg.RegressionNetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to derive multisig address: %w", err)
	}
	multisigAddressStr := p2wshAddr.EncodeAddress()

	// Get approved volunteers
	volunteers, err := s.repo.GetApprovedVolunteers(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get volunteers: %w", err)
	}
	if len(volunteers) == 0 {
		return nil, fmt.Errorf("no approved volunteers to pay out to")
	}

	// Fetch UTXO from Bitcoin RPC
	utxo, err := s.btcClient.FetchUTXO(multisigAddressStr, totalSats)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch UTXO for multisig address: %w", err)
	}

	// 6. Construct base PSBT
	tx := wire.NewMsgTx(wire.TxVersion)
	txHash, err := chainhash.NewHashFromStr(utxo.TxID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UTXO txid: %w", err)
	}
	outpoint := wire.NewOutPoint(txHash, utxo.Vout)
	txIn := wire.NewTxIn(outpoint, nil, nil)
	tx.AddTxIn(txIn)

	// Add outputs split equally among volunteers
	share := totalSats / int64(len(volunteers))
	for _, vol := range volunteers {
		dummyHash := sha256.Sum256([]byte(fmt.Sprintf("volunteer-%d-pamojabuild", vol.UserID)))
		pkScript, err := txscript.NewScriptBuilder().
			AddOp(txscript.OP_0).
			AddData(dummyHash[:20]). // P2WPKH dummy scriptPubKey
			Script()
		if err != nil {
			return nil, err
		}
		txOut := wire.NewTxOut(share, pkScript)
		tx.AddTxOut(txOut)
	}

	p2wshScriptPubKey, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).
		AddData(witnessScriptHash[:]).
		Script()
	if err != nil {
		return nil, err
	}

	p, err := psbt.NewFromUnsignedTx(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PSBT: %w", err)
	}

	p.Inputs[0].WitnessUtxo = &wire.TxOut{
		Value:    utxo.Amount,
		PkScript: p2wshScriptPubKey,
	}
	p.Inputs[0].WitnessScript = multisigWitnessScript

	var psbtBuf bytes.Buffer
	if err := p.Serialize(&psbtBuf); err != nil {
		return nil, fmt.Errorf("failed to serialize base PSBT: %w", err)
	}

	// 7. Store new payout_requests record
	pr := &models.PayoutRequest{
		TaskID:    taskID,
		TotalSats: totalSats,
		Status:    string(models.PayoutStatusPending),
		PSBT:      base64.StdEncoding.EncodeToString(psbtBuf.Bytes()),
	}

	err = s.repo.CreatePayoutRequest(ctx, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create payout request: %w", err)
	}

	// Update task status to pending_verification
	err = s.repo.UpdateTaskStatus(ctx, taskID, "pending_verification")
	if err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	return pr, nil
}

func (s *Service) SignPayoutRequest(ctx context.Context, id int64, userID int64) (*models.PayoutRequest, error) {
	// 1. Verify caller is keyholder
	keyholder, err := s.repo.GetKeyholderByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user is not a keyholder: %w", err)
	}

	// 2. Verify payout request exists and is pending
	pr, err := s.repo.GetPayoutRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if pr.Status != string(models.PayoutStatusPending) {
		return nil, fmt.Errorf("payout request status is %s, not pending", pr.Status)
	}

	// 3. Verify keyholder has not already signed/rejected
	alreadySigned, err := s.repo.HasKeyholderSigned(ctx, id, keyholder.ID)
	if err != nil {
		return nil, err
	}
	if alreadySigned {
		return nil, fmt.Errorf("keyholder has already signed or rejected this payout request")
	}

	// 4. Fetch the private key (WIF)
	wifIndex := keyholder.ID - 1
	if wifIndex < 0 || wifIndex >= int64(len(s.config.KeyholderKeys)) {
		return nil, fmt.Errorf("keyholder key index %d out of bounds", wifIndex)
	}
	wifStr := s.config.KeyholderKeys[wifIndex]
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode WIF: %w", err)
	}
	privKey := wif.PrivKey

	// 5. Construct the base PSBT and sign it
	p, err := psbt.NewFromRawBytes(strings.NewReader(pr.PSBT), true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base PSBT: %w", err)
	}

	tx := p.UnsignedTx
	prevOut := p.Inputs[0].WitnessUtxo
	witnessScript := p.Inputs[0].WitnessScript

	fetcher := txscript.NewMultiPrevOutFetcher(map[wire.OutPoint]*wire.TxOut{
		tx.TxIn[0].PreviousOutPoint: prevOut,
	})
	sigHashes := txscript.NewTxSigHashes(tx, fetcher)

	sighash, err := txscript.CalcWitnessSigHash(witnessScript, sigHashes, txscript.SigHashAll, tx, 0, prevOut.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate sighash: %w", err)
	}

	sig := ecdsa.Sign(privKey, sighash)
	sigWithHashType := append(sig.Serialize(), byte(txscript.SigHashAll))

	// Construct a partial PSBT representing only this signature
	partialP, err := psbt.NewFromUnsignedTx(tx)
	if err != nil {
		return nil, err
	}
	partialP.Inputs[0].WitnessUtxo = prevOut
	partialP.Inputs[0].WitnessScript = witnessScript
	partialP.Inputs[0].PartialSigs = []*psbt.PartialSig{
		{
			PubKey:    privKey.PubKey().SerializeCompressed(),
			Signature: sigWithHashType,
		},
	}

	var partialBuf bytes.Buffer
	if err := partialP.Serialize(&partialBuf); err != nil {
		return nil, fmt.Errorf("failed to serialize partial PSBT: %w", err)
	}
	partialPSBTHex := hex.EncodeToString(partialBuf.Bytes())

	// Store signature record with action = 'sign'
	sigRecord := &models.PayoutSignature{
		PayoutRequestID: pr.ID,
		KeyholderID:     keyholder.ID,
		Action:          "sign",
		PSBTPartial:     partialPSBTHex,
	}
	err = s.repo.SavePayoutSignature(ctx, sigRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to save payout signature: %w", err)
	}

	// 6. Check if threshold is met (3 signatures)
	sigs, err := s.repo.GetPayoutSignatures(ctx, id)
	if err != nil {
		return nil, err
	}

	signCount := 0
	for _, s := range sigs {
		if s.Action == "sign" {
			signCount++
		}
	}

	if signCount >= 3 {
		err = s.finalizeAndBroadcast(ctx, pr, sigs)
		if err != nil {
			log.Printf("Finalization or broadcast failed: %v", err)
		}
	}

	return s.repo.GetPayoutRequest(ctx, id)
}

func (s *Service) RejectPayoutRequest(ctx context.Context, id int64, userID int64) (*models.PayoutRequest, error) {
	// 1. Verify caller is keyholder
	keyholder, err := s.repo.GetKeyholderByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user is not a keyholder: %w", err)
	}

	// 2. Verify payout request exists and is pending
	pr, err := s.repo.GetPayoutRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if pr.Status != string(models.PayoutStatusPending) {
		return nil, fmt.Errorf("payout request status is %s, not pending", pr.Status)
	}

	// 3. Verify keyholder has not already signed/rejected
	alreadySigned, err := s.repo.HasKeyholderSigned(ctx, id, keyholder.ID)
	if err != nil {
		return nil, err
	}
	if alreadySigned {
		return nil, fmt.Errorf("keyholder has already signed or rejected this payout request")
	}

	// Store signature record with action = 'reject' and PSBTPartial = empty
	sigRecord := &models.PayoutSignature{
		PayoutRequestID: pr.ID,
		KeyholderID:     keyholder.ID,
		Action:          "reject",
		PSBTPartial:     "",
	}
	err = s.repo.SavePayoutSignature(ctx, sigRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to save payout rejection: %w", err)
	}

	// Check rejections count
	sigs, err := s.repo.GetPayoutSignatures(ctx, id)
	if err != nil {
		return nil, err
	}

	rejectCount := 0
	for _, s := range sigs {
		if s.Action == "reject" {
			rejectCount++
		}
	}

	if rejectCount >= 3 {
		// Update payout request to rejected
		err = s.repo.UpdatePayoutRequestStatus(ctx, id, string(models.PayoutStatusRejected))
		if err != nil {
			return nil, err
		}
		// Update task status back to open
		err = s.repo.UpdateTaskStatus(ctx, pr.TaskID, "open")
		if err != nil {
			return nil, err
		}
	}

	return s.repo.GetPayoutRequest(ctx, id)
}

func (s *Service) finalizeAndBroadcast(ctx context.Context, pr *models.PayoutRequest, sigs []*models.PayoutSignature) error {
	basePSBT, err := psbt.NewFromRawBytes(strings.NewReader(pr.PSBT), true)
	if err != nil {
		return fmt.Errorf("failed to parse base PSBT: %w", err)
	}

	// Merge all partial signatures into the base PSBT
	for _, sig := range sigs {
		if sig.Action != "sign" || sig.PSBTPartial == "" {
			continue
		}
		pBytes, err := hex.DecodeString(sig.PSBTPartial)
		if err != nil {
			return fmt.Errorf("failed to decode partial PSBT hex: %w", err)
		}
		pPart, err := psbt.NewFromRawBytes(bytes.NewReader(pBytes), false)
		if err != nil {
			return fmt.Errorf("failed to parse partial PSBT: %w", err)
		}

		for _, ps := range pPart.Inputs[0].PartialSigs {
			exists := false
			for _, existingSig := range basePSBT.Inputs[0].PartialSigs {
				if bytes.Equal(existingSig.PubKey, ps.PubKey) {
					exists = true
					break
				}
			}
			if !exists {
				basePSBT.Inputs[0].PartialSigs = append(basePSBT.Inputs[0].PartialSigs, ps)
			}
		}
	}

	// Load and sort keyholder public keys to build matching witness signature order
	keyholders, err := s.repo.GetKeyholders(ctx)
	if err != nil {
		return err
	}
	pubKeys := make([]*btcec.PublicKey, len(keyholders))
	for i, kh := range keyholders {
		b, err := hex.DecodeString(kh.PublicKey)
		if err != nil {
			return err
		}
		pk, err := btcec.ParsePubKey(b)
		if err != nil {
			return err
		}
		pubKeys[i] = pk
	}
	sort.Slice(pubKeys, func(i, j int) bool {
		return bytes.Compare(pubKeys[i].SerializeCompressed(), pubKeys[j].SerializeCompressed()) < 0
	})

	// Construct scriptWitness: OP_0 <sig1> <sig2> <sig3> <witnessScript>
	witness := wire.TxWitness{}
	witness = append(witness, nil) // OP_0/nil for multisig bug

	for _, pk := range pubKeys {
		pkCompressed := pk.SerializeCompressed()
		for _, sig := range basePSBT.Inputs[0].PartialSigs {
			if bytes.Equal(sig.PubKey, pkCompressed) {
				witness = append(witness, sig.Signature)
			}
		}
	}
	witness = append(witness, basePSBT.Inputs[0].WitnessScript)

	tx := basePSBT.UnsignedTx
	tx.TxIn[0].Witness = witness

	// Serialize finalized transaction
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return fmt.Errorf("failed to serialize transaction: %w", err)
	}
	txHex := hex.EncodeToString(buf.Bytes())

	// Broadcast the transaction to regtest LND node
	txID, err := s.lndClient.PublishTransaction(txHex)
	if err != nil {
		_ = s.repo.UpdatePayoutRequestStatus(ctx, pr.ID, "broadcast_failed")
		return fmt.Errorf("broadcast failed: %w", err)
	}

	// Update payout request, task, and volunteer statuses
	err = s.repo.UpdatePayoutRequestStatusAndTxID(ctx, pr.ID, string(models.PayoutStatusReleased), txID)
	if err != nil {
		return err
	}

	err = s.repo.UpdateTaskStatus(ctx, pr.TaskID, "completed")
	if err != nil {
		return err
	}

	// Fetch approved volunteers BEFORE updating their statuses to paid
	volunteers, err := s.repo.GetApprovedVolunteers(ctx, pr.TaskID)
	if err != nil {
		return err
	}

	err = s.repo.UpdateVolunteersStatus(ctx, pr.TaskID, "approved", "paid")
	if err != nil {
		return err
	}

	// Release payout off-chain via Lightning invoices
	for _, vol := range volunteers {
		if vol.PaymentRequest != "" {
			payErr := s.lndClient.PayInvoice(vol.PaymentRequest)
			if payErr != nil {
				log.Printf("Failed to pay invoice %s for volunteer %d: %v", vol.PaymentRequest, vol.ID, payErr)
				_ = s.repo.UpdateVolunteerStatus(ctx, vol.ID, "pay_failed")
			}
		}
	}

	return nil
}

func (s *Service) CheckGhostKeyholders(ctx context.Context) {
	cutoff := time.Now().Add(-72 * time.Hour).UTC().Format("2006-01-02 15:04:05")
	prs, err := s.repo.GetStalePayoutRequests(ctx, cutoff)
	if err != nil {
		log.Printf("Failed to fetch stale payout requests: %v", err)
		return
	}

	for _, pr := range prs {
		sigs, err := s.repo.GetPayoutSignatures(ctx, pr.ID)
		if err != nil {
			log.Printf("Failed to get signatures for stale request %d: %v", pr.ID, err)
			continue
		}

		signCount := 0
		for _, sig := range sigs {
			if sig.Action == "sign" {
				signCount++
			}
		}

		// If at least 2 keyholders signed, auto-release the payout request
		if signCount >= 2 {
			err = s.repo.UpdatePayoutRequestStatusAndTxID(ctx, pr.ID, string(models.PayoutStatusReleased), "auto-released")
			if err != nil {
				log.Printf("Failed to update status for auto-release request %d: %v", pr.ID, err)
				continue
			}

			err = s.repo.UpdateTaskStatus(ctx, pr.TaskID, "completed")
			if err != nil {
				log.Printf("Failed to complete task %d for auto-release: %v", pr.TaskID, err)
				continue
			}

			volunteers, err := s.repo.GetApprovedVolunteers(ctx, pr.TaskID)
			if err != nil {
				log.Printf("Failed to get volunteers for auto-release task %d: %v", pr.TaskID, err)
				continue
			}

			err = s.repo.UpdateVolunteersStatus(ctx, pr.TaskID, "approved", "paid")
			if err != nil {
				log.Printf("Failed to update volunteers status to paid for auto-release task %d: %v", pr.TaskID, err)
				continue
			}

			// Pay invoices
			for _, vol := range volunteers {
				if vol.PaymentRequest != "" {
					payErr := s.lndClient.PayInvoice(vol.PaymentRequest)
					if payErr != nil {
						log.Printf("Failed to pay invoice %s for volunteer %d in auto-release: %v", vol.PaymentRequest, vol.ID, payErr)
						_ = s.repo.UpdateVolunteerStatus(ctx, vol.ID, "pay_failed")
					}
				}
			}
		}
	}
}

func (s *Service) GetMultisigAddress(ctx context.Context) (string, error) {
	keyholders, err := s.repo.GetKeyholders(ctx)
	if err != nil {
		return "", err
	}
	pubKeys := make([]*btcec.PublicKey, len(keyholders))
	for i, kh := range keyholders {
		b, err := hex.DecodeString(kh.PublicKey)
		if err != nil {
			return "", err
		}
		pk, err := btcec.ParsePubKey(b)
		if err != nil {
			return "", err
		}
		pubKeys[i] = pk
	}
	sort.Slice(pubKeys, func(i, j int) bool {
		return bytes.Compare(pubKeys[i].SerializeCompressed(), pubKeys[j].SerializeCompressed()) < 0
	})
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_3)
	for _, pk := range pubKeys {
		builder.AddData(pk.SerializeCompressed())
	}
	builder.AddOp(txscript.OP_5)
	builder.AddOp(txscript.OP_CHECKMULTISIG)
	multisigWitnessScript, err := builder.Script()
	if err != nil {
		return "", err
	}
	witnessScriptHash := sha256.Sum256(multisigWitnessScript)
	p2wshAddr, err := btcutil.NewAddressWitnessScriptHash(witnessScriptHash[:], &chaincfg.RegressionNetParams)
	if err != nil {
		return "", err
	}
	return p2wshAddr.EncodeAddress(), nil
}
