import chromadb
from langchain_huggingface import HuggingFaceEmbeddings
import os

# Define the path to your ChromaDB instance
DB_PATH = "./legal_db" 
COLLECTION_NAME = "ethiopian_law"

print(f"--- Checking ChromaDB: {DB_PATH}/{COLLECTION_NAME} ---")

try:
    # Initialize ChromaDB client & collections
    client = chromadb.PersistentClient(path=DB_PATH)    
    collection = client.get_or_create_collection(name=COLLECTION_NAME)

    # 1. Check if the collection exists and its count
    count = collection.count()
    print(f"📊 Total documents in collection '{COLLECTION_NAME}': {count}")

    if count == 0:
        print("⚠️ Collection is empty. This is why you're getting no results.")
        print("   Please ensure 'ingestion_pipeline.py' was run successfully and populated this specific database.")
        print("   Also, verify that the 'legal_db' folder inside 'app/' is the one containing the ChromaDB data.")
    else:
        # 2. Try to retrieve a few documents by ID (example: first 5)
        print("\n--- Attempting to retrieve a few documents by ID ---")
        try:
            # Note: IDs are typically 'doc_0', 'doc_1', etc. from ingestion_pipeline.py
            sample_ids = [f"doc_{i}" for i in range(min(5, count))]
            retrieved_docs = collection.get(ids=sample_ids, include=['documents', 'metadatas'])
            if retrieved_docs['documents']:
                print(f"✅ Successfully retrieved {len(retrieved_docs['documents'])} sample documents:")
                for i, doc in enumerate(retrieved_docs['documents']):
                    print(f"   - Doc ID: {sample_ids[i]}")
                    print(f"     Content snippet: {doc[:100]}...") # Print first 100 chars
                    print(f"     Metadata: {retrieved_docs['metadatas'][i]}")
            else:
                print("❌ Could not retrieve documents by sample IDs. IDs might be different or collection is unreadable.")
        except Exception as e:
            print(f"❌ Error retrieving documents by ID: {e}")

        # 3. Perform a test query and print raw results
        print("\n--- Performing a test query: 'legal age of marriage' ---")
        test_query = "legal age of marriage"
        
        # Initialize the embedding model (must match the one used during ingestion)
        embeddings = HuggingFaceEmbeddings(model_name="sentence-transformers/all-MiniLM-L6-v2")
        query_vector = embeddings.embed_query(test_query)

        query_results = collection.query(
            query_embeddings=[query_vector],
            n_results=5, # Get top 5 results
            include=["documents", "metadatas", "distances"]
        )

        if query_results and query_results["documents"]:
            print(f"✅ Query results found for '{test_query}':")
            for i, (doc, meta, dist) in enumerate(zip(
                query_results["documents"], 
                query_results["metadatas"], 
                query_results["distances"]
            )):
                source_info = meta.get("source", "Unknown Source")
                article_info = meta.get("article_number", "N/A")
                print(f"   - Result {i+1}:")
                print(f"     Source: {source_info}, Article: {article_info}")
                print(f"     Distance: {dist:.4f}")
                print(f"     Content snippet: {doc[:100]}...")
        else:
            print(f"❌ No query results returned from ChromaDB for '{test_query}'.")

except Exception as e:
    print(f"🚨 An error occurred while trying to access ChromaDB: {e}")
    print("   This might indicate an issue with the ChromaDB installation or database files.")

print("\n--- ChromaDB Check Complete ---")
