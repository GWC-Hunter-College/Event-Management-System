package models

type Image struct {
	ImageID string `db:"id" json:"imageId"`
	Purpose string `db:"purpose" json:"purpose"`
	// S3 Object Key. i.e. the path from the root of the bucket to the file.
	ObjectKey string `db:"object_key" json:"objectKey"`
	Filename  string `db:"filename" json:"filename"`
	Type      string `db:"mimetype" json:"mimetype"`
	CreatedAt string `db:"created_at" json:"createdAt"`
}
