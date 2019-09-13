package product

import (
	"html/template"
	"time"
)

// Product is an item we sell.
type Product struct {
	ID          string    `db:"product_id" json:"id"`             // Unique identifier.
	Name        string    `db:"name" json:"name"`                 // Display name of the product.
	Cost        int       `db:"cost" json:"cost"`                 // Price for one item in cents.
	Quantity    int       `db:"quantity" json:"quantity"`         // Original number of items available.
	Sold        int       `db:"sold" json:"sold"`                 // Aggregate field showing number of items sold.
	Revenue     int       `db:"revenue" json:"revenue"`           // Aggregate field showing total cost of sold items.
	UserID      string    `db:"user_id" json:"user_id"`           // ID of the user who created the product.
	DateCreated time.Time `db:"date_created" json:"date_created"` // When the product was added.
	DateUpdated time.Time `db:"date_updated" json:"date_updated"` // When the product record was last modified.
}

// NameHTML fixes encoding issues.
func (p *Product) NameHTML() template.HTML {
	return template.HTML(p.Name)
}

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	Name     string `json:"name" validate:"required"`
	Cost     int    `json:"cost" validate:"required,gte=0"`
	Quantity int    `json:"quantity" validate:"gte=1"`
}

// UpdateProduct defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateProduct struct {
	Name     *string `json:"name"`
	Cost     *int    `json:"cost" validate:"omitempty,gte=0"`
	Quantity *int    `json:"quantity" validate:"omitempty,gte=1"`
}

// Sale represents one item of a transaction where some amount of a product was
// sold. Quantity is the number of units sold and Paid is the total price paid.
// Note that due to haggling the Paid value might not equal Quantity sold *
// Product cost.
type Sale struct {
	ID          string    `db:"sale_id" json:"id"`
	ProductID   string    `db:"product_id" json:"product_id"`
	Quantity    int       `db:"quantity" json:"quantity"`
	Paid        int       `db:"paid" json:"paid"`
	DateCreated time.Time `db:"date_created" json:"date_created"`
}

// NewSale is what we require from clients for recording new transactions.
type NewSale struct {
	Quantity int `json:"quantity" validate:"gte=0"`
	Paid     int `json:"paid" validate:"gte=0"`
}
