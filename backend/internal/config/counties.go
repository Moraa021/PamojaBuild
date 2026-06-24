package config

var KenyanCounties = []string{
	"Mombasa", "Kwale", "Kilifi", "Tana River", "Lamu", "Garissa", "Wajir", "Mandera",
	"Marsabit", "Isiolo", "Meru", "Tharaka-Nithi", "Embu", "Kitui", "Machakos",
	"Makueni", "Nyandarua", "Nyeri", "Kirinyaga", "Murang'a", "Kiambu", "Turkana",
	"West Pokot", "Samburu", "Trans-Nzoia", "Uasin Gishu", "Elgeyo-Marakwet", "Nandi",
	"Baringo", "Laikipia", "Nakuru", "Narok", "Kajiado", "Kericho", "Bomet",
	"Kakamega", "Vihiga", "Bungoma", "Busia", "Siaya", "Homa Bay", "Migori",
	"Kisii", "Nyamira", "Nairobi",
}

func IsValidCounty(county string) bool {
	for _, c := range KenyanCounties {
		if c == county {
			return true
		}
	}
	return false
}