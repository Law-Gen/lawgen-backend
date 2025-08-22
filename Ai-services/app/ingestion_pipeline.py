import os
import json
import chromadb
from langchain_huggingface import HuggingFaceEmbeddings
from typing import List, Dict, Any

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DB_PATH_ABS = os.path.join(SCRIPT_DIR, "legal_db")

# 1. Initialize ChromaDB client and collection
client = chromadb.PersistentClient(path=DB_PATH_ABS)
collection = client.get_or_create_collection(
    name="ethiopian_law",
    metadata={"hnsw:space": "cosine"}
)

# 2. Set up embedding model
embeddings_model = HuggingFaceEmbeddings(
    model_name="sentence-transformers/all-MiniLM-L6-v2"
)

# 3. Ingest Pre-processed JSON Documents
def ingest_documents_from_json(json_data: List[Dict[str, Any]], clear_existing: bool = False):
    """
    Ingests a list of pre-processed JSON document chunks into ChromaDB.
    Each dictionary in the list should contain 'content' and 'metadata'.
    
    Args:
        json_data (List[Dict[str, Any]]): A list of dictionaries, where each dict
                                          represents a document chunk with 'content'
                                          and 'metadata' fields.
        clear_existing (bool): If True, clears all existing documents in the collection
                               before ingesting new ones. Use with caution.
    """
    if clear_existing:
        print("🗑️ Clearing all existing document embeddings from ChromaDB...")
        try:
            collection.delete(where={})
            print("✅ All old embeddings cleared.")
        except Exception as e:
            print(f"⚠️ Could not clear collection: {e}. Proceeding without clearing.")

    existing_document_contents = {
        doc for doc in collection.get(include=['documents'])['documents']
    }
    
    num_added = 0
    next_id_counter = collection.count()

    for i, chunk in enumerate(json_data):
        content = chunk.get("content")
        metadata = chunk.get("metadata", {})

        if not content:
            print(f"⚠️ Skipping chunk {i} due to missing 'content' field.")
            continue
        
        if content in existing_document_contents:
            continue

        try:
            vector = embeddings_model.embed_documents([content])[0]
            collection.add(
                ids=[f"doc_{next_id_counter + num_added}"],
                documents=[content],
                embeddings=[vector],
                metadatas=[metadata]
            )
            existing_document_contents.add(content)
            num_added += 1
            if (num_added % 100) == 0:
                print(f"Ingested {num_added} documents so far...")

        except Exception as e:
            print(f"⚠️ Error indexing chunk {i}: {e}")
            
    print(f"📚 Successfully added {num_added} new chunks to ChromaDB.")
    print(f"📊 Total chunks in database: {collection.count()}")

if __name__ == "__main__":
    PROCESSED_DATA_DIR = os.path.join(SCRIPT_DIR, "..", "data")

    all_processed_chunks = []
    print(f"🔍 Looking for processed JSON files in '{PROCESSED_DATA_DIR}'...")

    for filename in os.listdir(PROCESSED_DATA_DIR):
        if filename.endswith(".json"):
            file_path = os.path.join(PROCESSED_DATA_DIR, filename)
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    if isinstance(data, list):
                        for item in data:
                            if "content" in item and "source" in item:
                                chunk_content = item["content"]
                                chunk_metadata = {k: v for k, v in item.items() if k != "content"}
                                
                                if 'topics' in chunk_metadata and isinstance(chunk_metadata['topics'], list):
                                    chunk_metadata['topics'] = ", ".join(chunk_metadata['topics'])
                                    
                                all_processed_chunks.append({"content": chunk_content, "metadata": chunk_metadata})
                            else:
                                print(f"⚠️ Skipping malformed item in {filename}: {item}. Missing 'content' or 'source'.")
                    else:
                        print(f"⚠️ Skipping {filename}: Expected a JSON array, but found a {type(data)}.")
            except json.JSONDecodeError as e:
                print(f"❌ Error decoding JSON from {filename}: {e}")
            except Exception as e:
                print(f"❌ An unexpected error occurred while processing {filename}: {e}")

    if all_processed_chunks:
        print(f"✅ Found {len(all_processed_chunks)} total chunks to process.")
        ingest_documents_from_json(all_processed_chunks, clear_existing=False)
    else:
        print("❌ No valid processed JSON documents found to ingest. Please check your 'data' directory and JSON structure.")

