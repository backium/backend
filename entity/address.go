package entity

type Address struct {
	Line1      string `bson:"line1"`
	Line2      string `bson:"line2"`
	District   string `bson:"district"`
	Province   string `bson:"province"`
	Department string `bson:"department"`
}
