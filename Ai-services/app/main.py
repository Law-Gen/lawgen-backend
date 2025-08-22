from fastapi import FastAPI
from pydantic import BaseModel
from app.qa_engine import ask_question
from typing import Dict, Any, List

# Initialize the FastAPI application
app = FastAPI(
    title="Ethiopian Legal AI Assistant",
    description="A RAG-powered AI assistant for querying Ethiopian legal documents (Constitution, Civil Code, Criminal Code)."
)

# Define the request model for the /ask endpoint
class QueryRequest(BaseModel):
    query: str 

# Define the response model for the /ask endpoint
class QueryResponse(BaseModel):
    response: str 
    references: List[Dict[str, str]] # List of source document references (title and summary)

@app.post("/ask", response_model=QueryResponse)
async def ask(req: QueryRequest):
    """
    Receives a legal question and returns an AI-generated answer
    along with references to the source legal documents.
    """
    try:
        # Call the core QA engine to get the response and references.
        result = ask_question(req.query)
        return result

    except Exception as e:
        print(f"Internal Server Error: {e}")
        return QueryResponse(
            response=f"⚠️ Internal error: {str(e)}. Please try again or rephrase your question.",
            references=[]
        )
@app.get("/")
async def read_root():
    return {"message": "Welcome to the Ethiopian Legal AI Assistant API! Send a POST request to /ask with your legal query."}

