package main

import "time"

type Category struct {
	Id          int64
	Name        string
	Description string
}

type Product struct {
	Id           int64
	Category     Category
	Model        string
	Manufacturer string
	Price        float64
	Quantity     int
	ImageUrl     string
	WarrantyDays int
}

type Characteristic struct {
	Id   int64
	Name string
	Unit string
}

type ProductCharacteristic struct {
	Id             int64
	ProductId      int64
	Characteristic Characteristic
	Value          string
}

type Customer struct {
	Id        int64
	FirstName string
	LastName  string
	Email     string
}

type Order struct {
	Id        int64
	Customer  Customer
	CreatedAt time.Time
	Address   string
}

type OrderItem struct {
	Id           int64
	OrderId      int64
	Product      Product
	Quantity     int
	PricePerItem float64
}
