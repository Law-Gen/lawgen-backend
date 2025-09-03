import logging
import time
import json
import asyncio
from typing import Literal

from fastapi import FastAPI, File, UploadFile, HTTPException, WebSocket, WebSocketDisconnect
from fastapi.responses import StreamingResponse, Response
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from engine import ask_question
from azure_utils import transcribe_speech, synthesize_speech, translate_text

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Ethiopian Legal AI Assistant",
    description="A RAG-powered AI assistant for querying Ethiopian legal documents with speech and translation capabilities."
)

# CORS middleware to allow all origins, i used this coz i am in development i will updated it once the frontend is deployed 
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# --- Pydantic Models for Request Bodies ---

class QueryRequest(BaseModel):
    #Request model for the /ask endpoint.
    query: str
    k: int = Field(default=5, ge=1, le=20, description="Number of relevant documents to return")

class TTSRequest(BaseModel):
    #Request model for the /text-to-speech endpoint.
    text: str
    language: Literal["en", "am"] = Field(default="en", description="Target language for speech synthesis (en or am)")
    
class TranslateRequest(BaseModel):
    #Request model for the new /translate endpoint
    text: str
    target_lang: Literal["en", "am"]

class TranslateResponse(BaseModel):
    #Response model for the new /translate endpoint.
    translated_text: str

# --- API Endpoints ---
@app.post("/translate", response_model=TranslateResponse, tags=["Translation"])
async def translate_text_endpoint(request: TranslateRequest):
    """
    Translates text between Amharic and English.
    
    The source language is automatically determined based on the target language.
    - If target_lang is 'en', the source is assumed to be 'am'.
    - If target_lang is 'am', the source is assumed to be 'en'.
    
    This endpoint utilizes the Azure Translator service.
    """
    try:
        translated_text = await translate_text(request.text, request.target_lang)
        return {"translated_text": translated_text}
    except Exception as e:
        logger.error(f"Translation API error: {e}")
        raise HTTPException(status_code=500, detail=f"An error occurred during translation: {str(e)}")


@app.get("/")
async def read_root():
    return {"message": "Welcome to the Ethiopian Legal AI Assistant API!"}

@app.get("/health")
async def health_check():
    return {"status": "healthy", "timestamp": time.time()}

@app.post("/ask")
async def ask_stream(req: QueryRequest):
    """
    Receives a query and streams back relevant legal document chunks.
    This is the core RAG functionality.
    """
    try:
        if not req.query.strip():
            raise HTTPException(status_code=400, detail="Query cannot be empty")
        
        logger.info(f"Received query: '{req.query[:100]}...' with k={req.k}")
        
        return StreamingResponse(
            ask_question(req.query, k=req.k),
            media_type="application/x-ndjson"
        )
    except Exception as e:
        logger.error(f"Error processing query: {str(e)}", exc_info=True)
        raise HTTPException(
            status_code=500, 
            detail="An error occurred while processing your request"
        )

@app.post("/speech-to-text/{language}")
async def speech_to_text(language: Literal["en", "am"], file: UploadFile = File(...)):
    #Accepts an audio file and transcribes it into text using Azure Speech service.
    if not file.content_type or not file.content_type.startswith("audio/"):
        raise HTTPException(status_code=400, detail="Invalid file type. Please upload an audio file.")

    audio_queue = asyncio.Queue()
    
    async def stream_audio_to_queue():
        try:
            while chunk := await file.read(1024):
                await audio_queue.put(chunk)
        finally:
            await audio_queue.put(None)

    streamer_task = asyncio.create_task(stream_audio_to_queue())

    try:
        logger.info(f"Starting transcription for language: {language}")
        transcribed_text = await transcribe_speech(audio_queue, language)
        return {"text": transcribed_text}
    except Exception as e:
        logger.error(f"Error during transcription: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Failed to transcribe audio: {str(e)}")
    finally:
        if not streamer_task.done():
            streamer_task.cancel()

@app.post("/text-to-speech")
async def text_to_speech(req: TTSRequest):
    #Converts text to speech and returns the audio as an MP3 file.
    try:
        logger.info(f"Synthesizing speech for language: {req.language}")
        audio_bytes = await synthesize_speech(req.text, req.language)
        return Response(content=audio_bytes, media_type="audio/mpeg")
    except Exception as e:
        logger.error(f"Error during speech synthesis: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Failed to synthesize speech: {str(e)}")

@app.websocket("/translate-stream")
async def translate_stream(websocket: WebSocket):
    """
    WebSocket endpoint for real-time translation between English and Amharic.
    Expects JSON: {"text": "...", "target_lang": "en|am"}
    Streams back a string with the translated text.
    """
    await websocket.accept()
    logger.info("Translation WebSocket connection established.")
    try:
        while True:
            data = await websocket.receive_text()
            try:
                payload = json.loads(data)
                text = payload.get("text")
                target_lang = payload.get("target_lang")

                if not text or not target_lang:
                    await websocket.send_text("Error: 'text' and 'target_lang' fields are required.")
                    continue
                
                if target_lang not in ["en", "am"]:
                    await websocket.send_text("Error: 'target_lang' must be either 'en' or 'am'.")
                    continue

                translated_text = await translate_text(text, target_lang)
                await websocket.send_text(translated_text)

            except json.JSONDecodeError:
                await websocket.send_text("Error: Invalid JSON format.")
            except Exception as e:
                logger.error(f"Error during WebSocket translation: {e}")
                await websocket.send_text(f"An error occurred: {e}")

    except WebSocketDisconnect:
        logger.info("Translation WebSocket connection closed.")
    except Exception as e:
        logger.error(f"An unexpected error occurred in the WebSocket: {e}")
        await websocket.close(code=1011, reason="An internal error occurred.")
@app.websocket("/translate-stream")
async def websocket_endpoint(websocket: WebSocket):
    """
    WebSocket endpoint for real-time text translation.
    
    Receives JSON payload with 'text' and 'target_lang' and sends 
    back a string with the translated text.
    """
    await websocket.accept()
    logger.info("Translation WebSocket connection established.")
    try:
        while True:
            data = await websocket.receive_text()
            try:
                payload = json.loads(data)
                text = payload.get("text")
                target_lang = payload.get("target_lang")

                if not text or not target_lang:
                    await websocket.send_text("Error: 'text' and 'target_lang' fields are required.")
                    continue
                
                if target_lang not in ["en", "am"]:
                    await websocket.send_text("Error: 'target_lang' must be either 'en' or 'am'.")
                    continue

                translated_text = await translate_text(text, target_lang)
                await websocket.send_text(translated_text)

            except json.JSONDecodeError:
                await websocket.send_text("Error: Invalid JSON format.")
            except Exception as e:
                logger.error(f"Error during WebSocket translation: {e}")
                await websocket.send_text(f"An error occurred: {e}")

    except WebSocketDisconnect:
        logger.info("Translation WebSocket connection closed.")
    except Exception as e:
        logger.error(f"An unexpected error occurred in the WebSocket: {e}")
        await websocket.close(code=1011, reason="An internal error occurred.")