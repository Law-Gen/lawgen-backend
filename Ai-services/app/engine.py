import os
import re
import chromadb
import json
import time
import logging
import os
from chromadb.utils import embedding_functions 
from dotenv import load_dotenv
from langchain.memory import ConversationBufferMemory
from dotenv import load_dotenv
from typing import List, Dict, Any, AsyncGenerator, Iterator
from pydantic import BaseModel, Field
from typing import Iterator, Dict, List, Any

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Load environment variables
load_dotenv()
google_api_key = os.getenv("GOOGLE_API_KEY") 

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DB_PATH_ABS = os.path.join(SCRIPT_DIR, "legal_db")

memory = ConversationBufferMemory(
    memory_key="chat_history",
    return_messages=True
)

RELEVANCE_THRESHOLD = 0.4

client = chromadb.PersistentClient(path=DB_PATH_ABS)
collection = client.get_or_create_collection(
    name="ethiopian_law",
    embedding_function=embedding_functions.SentenceTransformerEmbeddingFunction(
        model_name="sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
    )
)

class LegalResponse(BaseModel):
    content: str = Field(description="The content of the legal document chunk")
    metadata: Dict = Field(description="Metadata about the document chunk")
    score: float = Field(description="Relevance score (higher is more relevant)")

def filter_documents_by_relevance(documents, metadatas, distances, threshold=0.4):
    """Filter documents based on relevance score"""
    relevant_docs = []
    for i, (doc, meta, distance) in enumerate(zip(documents, metadatas, distances)):
        # Convert distance to similarity score (1 - distance)
        similarity = 1 - distance
        if similarity >= threshold:
            relevant_docs.append({
                "content": doc,
                "metadata": meta,
                "score": similarity
            })
    return relevant_docs

def get_relevant_documents(query: str, n_results: int = 5) -> List[Dict]:
    #Retrieve the k most relevant legal document chunks for a query.
    results = collection.query(
        query_texts=[query],
        n_results=n_results * 2,  # Get more results initially for filtering
        include=['documents', 'metadatas', 'distances']
    )
    relevant_docs = filter_documents_by_relevance(
        results["documents"][0],
        results["metadatas"][0],
        results["distances"][0],
        RELEVANCE_THRESHOLD
    )
    relevant_docs.sort(key=lambda x: x["score"], reverse=True)
    return relevant_docs[:n_results]

async def ask_question(query: str, k: int = 5) -> AsyncGenerator[str, None]:
    #Returns the k most relevant legal document chunks for the query in a simplified format.
    try:
        results = collection.query(
            query_texts=[query],
            n_results=k * 2,
            include=['documents', 'metadatas']
        )

        if not results or not results.get("documents"):
            yield json.dumps({
                "results": [],
                "message": "No relevant legal documents found for your query."
            })
            return

        # Process and filter relevant documents
        output_docs = []
        for doc, meta in zip(results["documents"][0], results["metadatas"][0]):
            # Extract only the necessary information
            formatted_doc = {
                "content": doc,
                "source": meta.get("source", "Unknown Source"),
                "article_number": meta.get("article_number", "N/A"),
                "topics": meta.get("topics", [])
            }
            output_docs.append(formatted_doc)

        output_docs = output_docs[:k]

        yield json.dumps({
            "results": output_docs,
            "message": f"Found {len(output_docs)} relevant legal documents."
        })

    except Exception as e:
        logger.error(f"Error in ask_question: {str(e)}")
        yield json.dumps({
            "results": [],
            "message": "An error occurred while processing your request."
        })

def preprocess_legal_document(content: str, metadata: Dict[str, Any]) -> Dict[str, Any]:
    #Preprocess legal documents before ingestion
    cleaned_content = re.sub(r'\s+', ' ', content).strip()
    article_match = re.search(r'Article\s+(\d+[A-ZaZ]?)', cleaned_content)
    article_number = article_match.group(1) if article_match else None
    enhanced_metadata = {
        **metadata,
        "article_number": article_number,
        "word_count": len(cleaned_content.split()),
        "processed_timestamp": time.time(),
        "language": "en"
    }
    return {
        "content": cleaned_content,
        "metadata": enhanced_metadata
    }

def extract_topics(content: str) -> List[str]:
    """Extract key topics from document content"""
    # Basic topic extraction based on key terms
    topics = set()
    topic_patterns = {
        "human_rights": r"human rights|fundamental rights|basic rights",
        "economic": r"economic|financial|monetary|commerce",
        "social": r"social|cultural|society|community",
        "political": r"political|government|democracy|election",
        "justice": r"justice|court|judicial|legal",
        "education": r"education|learning|teaching|school",
        "health": r"health|medical|healthcare",
    }
    
    for topic, pattern in topic_patterns.items():
        if re.search(pattern, content.lower()):
            topics.add(topic)
    
    return list(topics) if topics else ["general"]

def ask_question(query: str) -> Iterator[Dict[str, Any]]:
    """
    Synchronous wrapper for document retrieval
    Returns relevant documents with their references and stored topics
    """
    try:
        if not query or not isinstance(query, str):
            raise ValueError("Query must be a non-empty string")
            
        relevant_docs = get_relevant_documents(query, n_results=5)
        
        formatted_results = []
        for doc in relevant_docs:
            formatted_results.append({
                "content": doc["content"],
                "metadata": {
                    "source": doc["metadata"].get("source", ""),
                    "article_number": doc["metadata"].get("article_number", ""),
                    "topics": doc["metadata"].get("topics", "").split(", ") if doc["metadata"].get("topics") else ["general"]
                }
            })
        
        yield {
            "results": formatted_results,
            "message": f"Found {len(formatted_results)} relevant legal documents."
        }
        
    except Exception as e:
        logger.error(f"Error in ask_question: {str(e)}")
        yield {
            "results": [],
            "message": f"An error occurred: {str(e)}"
        }