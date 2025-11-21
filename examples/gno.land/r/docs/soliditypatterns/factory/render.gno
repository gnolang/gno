package factory

import (
	"strconv"
	"strings"
)

func Render(path string) string {
	content := `
# Factory Pattern

This pattern allows creating multiple instances of contracts dynamically. The factory contract serves as a template creator that can deploy new contract instances with different parameters.

---

## In Solidity

^^^solidity
contract Product {
    string public name;
    uint256 public price;
    address public owner;

    constructor(string memory _name, uint256 _price, address _owner) {
        name = _name;
        price = _price;
        owner = _owner;
    }

    function updatePrice(uint256 _newPrice) public {
        require(msg.sender == owner, "Only owner can update price");
        price = _newPrice;
    }
}

contract ProductFactory {
    Product[] public products;
    mapping(address => Product[]) public userProducts;

    event ProductCreated(address productAddress, string name, address owner);

    function createProduct(string memory _name, uint256 _price) public returns (address) {
        Product newProduct = new Product(_name, _price, msg.sender);
        products.push(newProduct);
        userProducts[msg.sender].push(newProduct);

        emit ProductCreated(address(newProduct), _name, msg.sender);
        return address(newProduct);
    }

    function getProductCount() public view returns (uint256) {
        return products.length;
    }

    function getUserProductCount(address user) public view returns (uint256) {
        return userProducts[user].length;
    }
}
^^^

#### Explanation

This Solidity implementation includes:

* A ^Product^ contract that represents individual products with name, price, and owner
* A constructor that initializes the product with the provided parameters
* An ^updatePrice^ function that allows only the owner to modify the price
* A ^ProductFactory^ contract that manages product creation
* A ^createProduct^ function that deploys new Product contracts using the ^new^ keyword
* Arrays and mappings to track all products and user-specific products
* Events to log when new products are created

---

## Gno Version

^^^go
package factory

import (
    "std"
    "strconv"
)

type Product struct {
    Name  string
    Price uint64
    Owner std.Address
    ID    int
}

var (
    products     []Product
    userProducts map[std.Address][]int // maps user address to product IDs
    nextID       int = 1
)

func init() {
    userProducts = make(map[std.Address][]int)
}

func CreateProduct(name string, price uint64) int {
    std.AssertOriginCall()
    owner := std.GetOrigCaller()

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
    caller := std.GetOrigCaller()

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
^^^

#### Key Differences

* In Gno, we use structs instead of separate contracts for products since Gno doesn't support dynamic contract deployment
* We maintain global slices and maps to track products and ownership relationships
* Product IDs are used instead of contract addresses for referencing products
* The ^std.AssertOriginCall()^ ensures functions are called directly, not from other contracts
* ^std.GetOrigCaller()^ gets the original transaction sender, similar to ^msg.sender^ in Solidity
* Error handling uses ^panic()^ instead of ^require()^ statements

---
`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}

func RenderDemo() string {
	content := "\n## Live Demo\n\n"

	content += "**Factory Statistics:**\n"
	content += "- Total Products: **" + strconv.Itoa(GetProductCount()) + "**\n\n"

	if len(products) > 0 {
		content += "**All Products:**\n"
		for _, product := range products {
			content += "- Product #" + strconv.Itoa(product.ID) + ": **" + product.Name + "** - " + strconv.FormatUint(product.Price, 10) + " units\n"
			content += "  - Owner: `" + string(product.Owner) + "`\n"
		}
		content += "\n"
	} else {
		content += "*No products created yet.*\n\n"
	}

	content += "**Actions:**\n"
	content += "[Create Laptop Product](factory$help&func=CreateProduct&args=Laptop,1200)\n\n"
	content += "[Create Phone Product](factory$help&func=CreateProduct&args=Phone,800)\n\n"
	content += "[Create Tablet Product](factory$help&func=CreateProduct&args=Tablet,600)\n\n"

	if len(products) > 0 {
		firstProduct := products[0]
		newPrice := firstProduct.Price + 100
		content += "[Update Product #" + strconv.Itoa(firstProduct.ID) + " Price](factory$help&func=UpdateProductPrice&args=" + strconv.Itoa(firstProduct.ID) + "," + strconv.FormatUint(newPrice, 10) + ")\n\n"
	}

	content += "**Query Functions:**\n"
	content += "[Get Product Count](factory$help&func=GetProductCount)\n\n"

	if len(products) > 0 {
		firstProduct := products[0]
		content += "[Get Product #" + strconv.Itoa(firstProduct.ID) + " Details](factory$help&func=GetProduct&args=" + strconv.Itoa(firstProduct.ID) + ")\n\n"
	}

	return content
}
