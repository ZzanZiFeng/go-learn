	for n := range square(gen(1, 2, 3, 4, 5)) {
		fmt.Println("Squared:", n)
	}
	// Output:
	// Squared: 1
	// Squared: 4
	// Squared: 9
	// Squared: 16
	// Squared: 25

	fmt.Println("\n=== Channels Demo Complete ===")
}