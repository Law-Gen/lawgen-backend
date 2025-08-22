import os
import re
import chromadb
from dotenv import load_dotenv
from langchain_google_genai import ChatGoogleGenerativeAI 
from langchain.prompts import PromptTemplate
from langchain_core.runnables import RunnableMap
from langchain_core.output_parsers import StrOutputParser
from langchain.memory import ConversationBufferMemory
from langchain_huggingface import HuggingFaceEmbeddings
from typing import List, Dict, Any, Tuple, Optional

load_dotenv()
google_api_key = os.getenv("GOOGLE_API_KEY") 

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DB_PATH_ABS = os.path.join(SCRIPT_DIR, "legal_db")

#Initialize the LLM for Google Gemini
llm = ChatGoogleGenerativeAI(
    google_api_key=google_api_key,
    model="gemini-2.5-flash", 
    temperature=0.3
)

# Session memory for conversational context
memory = ConversationBufferMemory(memory_key="chat_history", return_messages=True)

# Define the relevance threshold for retrieved documents.
RELEVANCE_THRESHOLD = 1.1 

client = chromadb.PersistentClient(path=DB_PATH_ABS) 
collection = client.get_or_create_collection("ethiopian_law")

# Initialize the HuggingFaceEmbeddings object globally
embeddings = HuggingFaceEmbeddings(
    model_name="sentence-transformers/all-MiniLM-L6-v2"
)

# Structured conversational prompt for the LLM.
prompt = PromptTemplate(
    input_variables=["context", "question", "chat_history"],
    template="""
You are an expert legal AI assistant specifically trained on the Ethiopian Constitution, Civil Code, Criminal Code, and other Ethiopian legal documents.
Your primary directive is to provide accurate and relevant legal information.

Instructions:
1.  **CRITICAL: STRICTLY USE ONLY THE "Legal Context" PROVIDED.** Do NOT use any external knowledge. If the provided context is insufficient, state the fallback message.
2.  **VERY IMPORTANT: Synthesize and Answer FROM RELEVANT CONTEXT.** If ANY part of the "Legal Context" contains information that helps answer the "User Question", you MUST synthesize a comprehensive answer from that information.
    * **Interpret related terms:** Understand that a query about "tax evasion" is related to context discussing "duties and taxes not paid," "customs duties," "fiscal offenses," or similar concepts. Do not require an exact keyword match to consider context relevant.
    * **Prioritize direct relevance:** Even if multiple documents are retrieved, focus primarily on the most directly applicable articles for your answer.
3.  **Fallback ONLY if TRULY no answer:** If the "Legal Context" is completely empty, or if AFTER CAREFUL AND THOROUGH REVIEW you determine that ABSOLUTELY NO part of the "Legal Context" provides information to answer the "User Question", you MUST respond with the following exact phrase and NOTHING ELSE:
        "I can only provide information based on the provided legal documents (Ethiopian Constitution, Civil Code, and Criminal Code, along with other relevant Ethiopian laws). Your question appears to be outside my knowledge base. Please ask about a relevant topic."
    **DO NOT add any additional notes, explanations, or justifications to this fallback message.**
4.  **Thought Process (Internal consideration before answering):**
    * Analyze the "User Question" for core legal concepts.
    * Thoroughly examine ALL provided "Legal Context" documents to identify all relevant articles or sections, even if the phrasing differs from the question.
    * Interpret the legal meaning of these relevant sections specifically in relation to the question.
    * Integrate information from all truly relevant parts of the context to form a complete, accurate, and coherent answer.
5.  **Citations:**
    * Cite specific article numbers (e.g., "Article 1042") ONLY if they are explicitly mentioned and relevant within the provided "Legal Context" for the points you are making.
    * Always use the full source document title as it appears in the references (e.g., "THE CRIMINAL CODE OF THE FEDERAL DEMOCRATIC REPUBLIC OF ETHIOPIA - Article 352").
    * Do NOT invent or guess article numbers or source titles.
6.  **Tone and Style:** Maintain an insightful, warm, and professional tone. Your response should be cohesive, human, and conversational.
7.  **Response Structure:**
    * Start with a topic-aware opening directly addressing the question.
    * Clearly explain the law based on the context.
    * Conclude with a concise legal takeaway or implication.
    * End with a subtle, natural follow-up prompt.
    * Include a comprehensive source reference (e.g., “📘 Source: Article 352, THE CRIMINAL CODE OF THE FEDERAL DEMOCRATIC REPUBLIC OF ETHIOPIA”) at the very end, combining all *used* relevant sources.

---

Chat History:
{chat_history}

Legal Context:
{context}

User Question: {question}

Answer:
"""
)

# Create the qa_chain globally after prompt and llm are defined
qa_chain = (
    RunnableMap({
        "context": lambda x: x["context"],
        "question": lambda x: x["question"],
        "chat_history": lambda x: x.get("chat_history", "")
    })
    | prompt
    | llm
    | StrOutputParser()
)


def search_legal_docs(query: str, top_k: int = 5, max_chars: int = 4000) -> Tuple[Optional[str], Optional[Dict]]:
    """
    Performs a similarity search in ChromaDB and filters results based on relevance.
    Leverages metadata for potential future filtering enhancements (e.g., by topic).
    
    Args:
        query (str): The user's question.
        top_k (int): The number of top similar documents to retrieve from the vector store.
        max_chars (int): Maximum character length for the combined context sent to the LLM.
        
    Returns:
        Tuple[Optional[str], Optional[Dict]]: A tuple containing:
                                                - The combined relevant context string (or None if no relevant docs).
                                                - A dictionary of grouped metadata (or None).
    """
    query_vector = embeddings.embed_query(query)
    
    # Perform vector similarity search and explicitly request distances and metadatas
    results = collection.query(
        query_embeddings=[query_vector],
        n_results=top_k,
        include=["documents", "metadatas", "distances"]
    )

    # DEBUGGING: Print all retrieved distances to see what ChromaDB is returning
    print(f"\n🔍 Query: '{query}'")
    if results and results["distances"] and results["distances"][0]:
        print(f"   Raw distances from ChromaDB: {results['distances'][0]}")
    else:
        print("   No distances retrieved from ChromaDB.")


    relevant_grouped_docs = {}
    has_relevant_doc = False

    if results and results["documents"] and results["documents"][0]:
        for i, (doc_content, doc_meta, doc_distance) in enumerate(zip(
            results["documents"][0], results["metadatas"][0], results["distances"][0]
        )):
            # Only include documents that meet the relevance threshold
            if doc_distance <= RELEVANCE_THRESHOLD:
                has_relevant_doc = True
                
                source_doc = doc_meta.get("source", "Unknown Source")
                article_num = doc_meta.get("article_number", "")
                summary_text = doc_meta.get("summary", "")

                title_key = f"{source_doc} - Article {article_num}" if article_num else source_doc
                
                if title_key not in relevant_grouped_docs:
                    relevant_grouped_docs[title_key] = {
                        "content": [], 
                        "summary": summary_text
                    }
                relevant_grouped_docs[title_key]["content"].append(doc_content.strip())
                # DEBUGGING: Confirm which documents are being included
                print(f"   ✅ Including doc {i} (Distance: {doc_distance:.4f}) - {title_key}")
            else:
                # DEBUGGING: Confirm which documents are being excluded
                source_doc = doc_meta.get("source", "Unknown Source")
                article_num = doc_meta.get("article_number", "")
                title_key = f"{source_doc} - Article {article_num}" if article_num else source_doc
                print(f"   ❌ Excluding doc {i} (Distance: {doc_distance:.4f} > Threshold: {RELEVANCE_THRESHOLD}) - {title_key}")
    
    if not has_relevant_doc:
        print(f"   🛑 No relevant documents found after filtering with threshold {RELEVANCE_THRESHOLD}.")
        return None, None # No relevant documents found based on threshold

    # Construct the context string for the LLM
    sorted_titles = sorted(
        relevant_grouped_docs.keys(),
        key=lambda x: int(re.findall(r'\d+', x)[0]) if re.search(r'\d+', x) else 9999
    )

    context_parts = []
    for title in sorted_titles:
        summary_from_meta = relevant_grouped_docs[title].get('summary', '')
        summary_line = f"Summary: {summary_from_meta}\n" if summary_from_meta else ""
        context_parts.append(f"Source: {title}\n{summary_line}Content:\n" + "\n".join(relevant_grouped_docs[title]["content"]))

    full_context = "\n\n".join(context_parts)

    return full_context[:max_chars], relevant_grouped_docs


def ask_question(query: str) -> Dict[str, Any]:
    """
    Main function to answer a legal question using RAG.
    
    Args:
        query (str): The user's question.
        
    Returns:
        Dict[str, Any]: A dictionary containing the 'response' from the LLM
                        and 'references' (list of dicts) to the source documents.
    """
    # Step 1: Retrieve relevant context from the vector database
    context, grouped_metadata = search_legal_docs(query)
    
    # If no relevant context is found, return the predefined message.
    if context is None:
        return {
            "response": "I can only provide information based on the provided legal documents (Ethiopian Constitution, Civil Code, and Criminal Code, along with other relevant Ethiopian laws). Your question appears to be outside my knowledge base. Please ask about a relevant topic.",
            "references": []
        }

    # Step 2: Load chat history for conversational context
    chat_history = memory.load_memory_variables({}).get("chat_history", [])
    
    # Step 3: Invoke the LLM with the prompt, context, and chat history
    llm_response = qa_chain.invoke({
        "context": context,
        "question": query,
        "chat_history": chat_history
    })
    
    # Step 4: Save the current conversation turn to memory for future context
    memory.save_context({"input": query}, {"output": llm_response})
    
    # Step 5: Prepare references for the frontend display
    references = []
    if grouped_metadata:
        for key in sorted(
            grouped_metadata.keys(),
            key=lambda x: int(re.findall(r'\d+', x)[0]) if re.search(r'\d+', x) else 9999
        ):
            summary = grouped_metadata[key].get("summary", "")
            references.append({"title": key, "summary": summary})

    return {
        "response": llm_response,
        "references": references
    }

