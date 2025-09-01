from fastapi import FastAPI
from fastapi.responses import StreamingResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
import grpc
import logging
import time
import json
from concurrent import futures

from app.engine import ask_question
from app.legal_assistant_pb2 import (
    QuestionResponse, 
    HealthCheckResponse
)
from app.legal_assistant_pb2_grpc import (
    LegalAssistantServicer, 
    add_LegalAssistantServicer_to_server
)

app = FastAPI(
    title="Ethiopian Legal AI Assistant",
    description="A RAG-powered AI assistant for querying Ethiopian legal documents (Constitution, Civil Code, Criminal Code)."
)

#CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class QueryRequest(BaseModel):
    query: str
    k: int = Field(default=5, ge=1, le=20, description="Number of relevant documents to return")

@app.post("/ask")
async def ask_stream(req: QueryRequest):
    start_time = time.time()
    try:
        if not req.query.strip():
            return {"error": "Query cannot be empty"}            
        logger.info(f"Received query: {req.query[:100]}... with k={req.k}")
        
        return StreamingResponse(
            ask_question(req.query, k=req.k),
            media_type="application/json"
        )
    except Exception as e:
        logger.error(f"Error processing query: {str(e)}", exc_info=True)
        return {
            "error": "An error occurred while processing your request",
            "details": str(e)
        }

@app.get("/")
async def read_root():
    return {"message": "Welcome to the Ethiopian Legal AI Assistant API! Send a POST request to /ask with your query."}

@app.get("/health")
async def health_check():
    return {"status": "healthy", "timestamp": time.time()}

class LegalAssistantService(LegalAssistantServicer):
    def AskQuestion(self, request, context):
        try:
            if not request.query.strip():
                context.abort(grpc.StatusCode.INVALID_ARGUMENT, "Query cannot be empty")
            
            k = max(1, min(request.k, 20)) if request.k else 5
            logger.info(f"Received query: {request.query[:100]}... with k={k}")
            
            for response in ask_question(request.query, k=k):
                result = json.loads(response)
                yield QuestionResponse(
                    text=json.dumps(result, indent=2),
                    references=[doc.get("source", "") for doc in result.get("results", [])]
                )
                
        except Exception as e:
            logger.error(f"Error processing query: {str(e)}")
            context.abort(grpc.StatusCode.INTERNAL, f"An error occurred: {str(e)}")

    def HealthCheck(self, request, context):
        return HealthCheckResponse(
            status="healthy",
            timestamp=time.time()
        )

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    add_LegalAssistantServicer_to_server(LegalAssistantService(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    logger.info("gRPC Server started on port 50051")
    server.wait_for_termination()

if __name__ == '__main__':
    serve()