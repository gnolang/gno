package factory

import (
    "std"
)

type Product struct {
    Name  string
    Price uint64
    Owner std.Address
    ID    int
}

var (
    products     []Product
    userProducts map[std.Address][]int
    nextID       int = 1
)

func init() {
    userProducts = make(map[std.Address][]int)
}

func CreateProduct(name string, price uint64) int {
    std.AssertOriginCall()
    owner := std.OriginCaller()

    product := Product{
        Name:  name,
        Price: price,
        Owner: owner,
        ID:    nextID,
    }

    products = append(products, product)
    userProducts[owner] = append(userProducts[owner], nextID)

    nextID++
    return product.ID
}

func UpdateProductPrice(productID int, newPrice uint64) {
    std.AssertOriginCall()
    caller := std.OriginCaller()

    for i, product := range products {
        if product.ID == productID {
            if product.Owner != caller {
                panic("Only owner can update price")
            }
            products[i].Price = newPrice
            return
        }
    }
    panic("Product not found")
}

func GetProduct(productID int) (string, uint64, std.Address) {
    for _, product := range products {
        if product.ID == productID {
            return product.Name, product.Price, product.Owner
        }
    }
    panic("Product not found")
}


func GetProductCount() int {
    return len(products)
}


func GetUserProductCount(user std.Address) int {
    return len(userProducts[user])
}


func GetUserProducts(user std.Address) []int {
    return userProducts[user]
}
