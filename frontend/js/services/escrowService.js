import escrowApi  from '../api/escrowApi.js';
import { signWithPrivateKey, toBytes } from '../utils/crypto.js';

const escrowService = {
  async buildCoSignPayload({ privateKey, trusteePubKeyHex, layer1PsbtFragment, volunteerInvoice }) {
    const layer2Sig = await signWithPrivateKey(privateKey, toBytes(volunteerInvoice));

    return {
      trustee_public_key_hex:          trusteePubKeyHex,
      layer1_psbt_signature_fragment:  layer1PsbtFragment,
      layer2_web_crypto_signature:     layer2Sig,
    };
  },

  async submitSignature(slug, signParams) {
    const payload = await this.buildCoSignPayload(signParams);
    return escrowApi.submitCoSignatures(slug, payload);
  },

  signaturesRemaining(collected, threshold = 3) {
    return Math.max(0, threshold - collected);
  },
};

export default escrowService;
