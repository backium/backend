package repository

type addressRecord struct {
	Line1      string `bson:"line1,omitempty"`
	Line2      string `bson:"line2,omitempty"`
	District   string `bson:"district,omitempty"`
	Province   string `bson:"province,omitempty"`
	Department string `bson:"department,omitempty"`
}
