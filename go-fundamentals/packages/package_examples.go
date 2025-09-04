package packages

import (
	"fmt"
	// Import the local helper package using its module path.
	"go-study/go-fundamentals/packages/helper"
)

func DemonstratePackages() {
	fmt.Println("\n--- Package Demonstrations ---")

	// Call the exported function from the helper package.
	result := helper.PublicFunction("hello from helper package")
	fmt.Println("Result from helper.PublicFunction:", result)

	// The following line would cause a compile error because privateFunction is not exported:
	// helper.privateFunction()
	fmt.Println("Note: We cannot call private functions from other packages.")
}
