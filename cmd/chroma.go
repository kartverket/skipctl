package cmd

import (
	"context"
	"fmt"

	"github.com/kartverket/skipctl/pkg/vectordb"
	"github.com/spf13/cobra"
)

var (
	chromaURL        string
	chromaCollection string
)

func init() {
	chromaCmd.PersistentFlags().StringVar(&chromaURL, "chroma-url", "http://localhost:8000", "Chroma server URL")
	chromaCmd.PersistentFlags().StringVar(&chromaCollection, "collection", "argokit", "Collection name")

	chromaCmd.AddCommand(chromaIndexCmd)
	chromaCmd.AddCommand(chromaSearchCmd)

	chromaSearchCmd.Flags().IntP("num-results", "n", 3, "Number of results to return")

	rootCmd.AddCommand(chromaCmd)
}

var chromaCmd = &cobra.Command{
	Use:   "chroma",
	Short: "Manage ArgoKit documentation in Chroma vector database",
	Long: `Manage your ArgoKit documentation stored in Chroma vector database.

Use these commands to index ArgoKit documentation and search for patterns.`,
}

var chromaIndexCmd = &cobra.Command{
	Use:   "index [directory]",
	Short: "Index ArgoKit documentation into Chroma",
	Long: `Index ArgoKit documentation (.md and .jsonnet files) from a directory into Chroma.

This command recursively scans the directory for .md and .jsonnet files and stores
them in the Chroma vector database for semantic search.

Examples:
  # Index ArgoKit knowledge base
  skipctl chroma index ./argokit-knowledge

  # Index into specific collection
  skipctl chroma index ./team-docs --collection team-argokit

  # Use custom Chroma URL
  skipctl chroma index ./docs --chroma-url http://chroma.mycompany.com:8000`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docsPath := args[0]
		ctx := context.Background()

		client := vectordb.NewLocalChromaClient(chromaURL)

		// Check if Chroma is available
		if !client.IsAvailable(ctx) {
			return fmt.Errorf("Chroma server not available at %s\n\nStart with: docker-compose -f docker-compose.chroma.yml up -d", chromaURL)
		}

		fmt.Printf("Chroma Server: %s\n", chromaURL)
		fmt.Printf("Documentation: %s\n", docsPath)
		fmt.Printf("Collection: %s\n\n", chromaCollection)

		// Get or create collection
		collection, err := client.GetOrCreateCollection(ctx, chromaCollection)
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}

		fmt.Printf("Collection ready (ID: %s)\n\n", collection.ID)

		// Index documents
		fmt.Println("Indexing documents...")
		count, err := client.LoadDocumentsFromDirectory(ctx, collection.Name, docsPath)
		if err != nil {
			return fmt.Errorf("failed to index documents: %w", err)
		}

		fmt.Printf("\nIndexed %d documents into collection '%s'\n\n", count, chromaCollection)
		fmt.Println("Now you can use refactor with this collection:")
		fmt.Printf("  skipctl refactor deployment.yaml --collection %s -o app.jsonnet\n", chromaCollection)

		return nil
	},
}

var chromaSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for ArgoKit patterns by semantic similarity",
	Long: `Search for ArgoKit documentation and examples using semantic similarity.

This helps you verify what patterns are available in your vector database.

Examples:
  skipctl chroma search "kubernetes deployment"
  skipctl chroma search "database statefulset" --collection argokit -n 5`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		numResults, _ := cmd.Flags().GetInt("num-results")
		ctx := context.Background()

		client := vectordb.NewLocalChromaClient(chromaURL)

		if !client.IsAvailable(ctx) {
			return fmt.Errorf("Chroma server not available at %s", chromaURL)
		}

		fmt.Printf("Searching: \"%s\"\n", query)
		fmt.Printf("Collection: %s\n\n", chromaCollection)

		results, err := client.Query(ctx, chromaCollection, query, numResults)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results) == 0 {
			fmt.Printf("No results found in collection '%s'\n\n", chromaCollection)
			fmt.Println("Index documentation first:")
			fmt.Printf("  skipctl chroma index ./argokit-knowledge --collection %s\n", chromaCollection)
			return nil
		}

		fmt.Printf("Found %d result(s):\n\n", len(results))
		for i, doc := range results {
			fmt.Printf("--- Result %d ---\n", i+1)
			fmt.Printf("%s\n\n", doc)
		}

		return nil
	},
}
