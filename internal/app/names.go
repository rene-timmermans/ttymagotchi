package app

import (
	crand "crypto/rand"
	"math/big"
)

// Top 100 first given names by total carriers in the Netherlands (2022).
// Source: Nederlandse Voornamenbank, Meertens Instituut KNAW.
// https://nvb.meertens.knaw.nl/downloads/Top_eerste_voornamen_NL_2022.zip
// Female and male carrier counts are combined for names shared by both groups.
var tamagotchiNames = []string{
	"Maria",
	"Johannes",
	"Johanna",
	"Jan",
	"Anna",
	"Cornelis",
	"Hendrik",
	"Elisabeth",
	"Cornelia",
	"Willem",
	"Wilhelmina",
	"Petrus",
	"Catharina",
	"Pieter",
	"Gerrit",
	"Robert",
	"Peter",
	"Hendrika",
	"Adriana",
	"Wilhelmus",
	"Petronella",
	"Jacobus",
	"Antonius",
	"Gerardus",
	"Adrianus",
	"Franciscus",
	"Jeroen",
	"Hendrikus",
	"Ronald",
	"Johan",
	"Thomas",
	"Jacob",
	"Yvonne",
	"Marinus",
	"Dirk",
	"Monique",
	"Margaretha",
	"Anne",
	"Martinus",
	"Dennis",
	"Theodorus",
	"Richard",
	"Marcel",
	"Jacoba",
	"Helena",
	"Mark",
	"Esther",
	"Henricus",
	"Tim",
	"Patrick",
	"Ingrid",
	"Geertruida",
	"Robin",
	"René",
	"Linda",
	"Martijn",
	"Sandra",
	"Paul",
	"Erik",
	"Sanne",
	"Nicolaas",
	"Albert",
	"Alexander",
	"Maarten",
	"Daan",
	"Michael",
	"Mohamed",
	"Daniël",
	"Laura",
	"Kim",
	"Marco",
	"Sander",
	"Christina",
	"Arie",
	"Wouter",
	"Eva",
	"Elizabeth",
	"Niels",
	"Bart",
	"Antonia",
	"Lucas",
	"Frank",
	"Tom",
	"Bram",
	"Iris",
	"Emma",
	"Lisa",
	"Kevin",
	"Marianne",
	"Ruben",
	"Karin",
	"Jolanda",
	"Lars",
	"Paulus",
	"Theodora",
	"David",
	"Thijs",
	"Bernardus",
	"Michel",
	"Julia",
}

func randomName() string {
	if len(tamagotchiNames) == 0 {
		return "Olaf"
	}
	upper := big.NewInt(int64(len(tamagotchiNames)))
	n, err := crand.Int(crand.Reader, upper)
	if err != nil {
		return tamagotchiNames[0]
	}
	return tamagotchiNames[n.Int64()]
}
