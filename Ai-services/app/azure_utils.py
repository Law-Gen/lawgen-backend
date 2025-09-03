import os
import uuid
import asyncio
import logging
import aiohttp
import azure.cognitiveservices.speech as speechsdk
from azure.storage.blob.aio import BlobServiceClient
from azure.storage.blob import BlobSasPermissions, generate_blob_sas
from datetime import datetime, timedelta
from pathlib import Path
from dotenv import load_dotenv

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Load environment variables
ROOT_DIR = Path(__file__).parent.parent
load_dotenv(ROOT_DIR / ".env")

# Azure credentials
STORAGE_CONN_STRING = os.getenv("AZURE_STORAGE_CONNECTION_STRING")
STORAGE_ACCOUNT = os.getenv("AZURE_STORAGE_ACCOUNT") 
STORAGE_ACCOUNT_KEY = os.getenv("AZURE_STORAGE_ACCOUNT_KEY")

TRANSLATOR_KEY = os.getenv("TRANSLATOR_KEY")
TRANSLATOR_ENDPOINT = os.getenv("TRANSLATOR_ENDPOINT", "https://api.cognitive.microsofttranslator.com/")
TRANSLATOR_REGION = os.getenv("TRANSLATOR_REGION")

# New: Separate credentials for STT and TTS services
STT_SPEECH_KEY = os.getenv("STT_SPEECH_KEY")
STT_SPEECH_REGION = os.getenv("STT_SPEECH_REGION")
TTS_SPEECH_KEY = os.getenv("TTS_SPEECH_KEY")
TTS_SPEECH_REGION = os.getenv("TTS_SPEECH_REGION")

# Assert that all required environment variables are set
assert STT_SPEECH_KEY, "STT_SPEECH_KEY not set"
assert STT_SPEECH_REGION, "STT_SPEECH_REGION not set"
assert TTS_SPEECH_KEY, "TTS_SPEECH_KEY not set"
assert TTS_SPEECH_REGION, "TTS_SPEECH_REGION not set"

# --- Azure Speech SDK Utils ---

async def transcribe_speech(audio_queue: asyncio.Queue, language: str) -> str:
    """
    Transcribes audio from an asynchronous queue of data chunks.
    This function uses a PushAudioInputStream to feed audio data to the recognizer in real-time.
    """
    try:
        speech_config = speechsdk.SpeechConfig(subscription=STT_SPEECH_KEY, region=STT_SPEECH_REGION)
        lang_code = "en-US" if language == "en" else "am-ET"
        speech_config.speech_recognition_language = lang_code
        
        # Create a PushAudioInputStream to feed the in-memory audio data
        push_stream = speechsdk.audio.PushAudioInputStream()
        audio_config = speechsdk.audio.AudioConfig(stream=push_stream)
        recognizer = speechsdk.SpeechRecognizer(
            speech_config=speech_config, 
            audio_config=audio_config
        )

        # Asynchronously read from the queue and write to the PushAudioInputStream
        async def feed_stream():
            while True:
                chunk = await audio_queue.get()
                if chunk is None:  # End of stream signal
                    push_stream.close()
                    break
                push_stream.write(chunk)

        feed_task = asyncio.create_task(feed_stream())
        
        # Recognize once and return the recognized text
        result = await asyncio.get_event_loop().run_in_executor(None, recognizer.recognize_once)
        
        await feed_task
        
        if result.reason == speechsdk.ResultReason.RecognizedSpeech:
            return result.text
        elif result.reason == speechsdk.ResultReason.NoMatch:
            logger.warning("No speech could be recognized.")
            return ""
        elif result.reason == speechsdk.ResultReason.Canceled:
            cancellation_details = result.cancellation_details
            logger.error(f"Speech recognition canceled: {cancellation_details.reason}")
            if cancellation_details.reason == speechsdk.CancellationReason.Error:
                logger.error(f"Error details: {cancellation_details.error_details}")
            raise Exception(f"Speech recognition canceled: {cancellation_details.reason}")
    except Exception as e:
        logger.error(f"Transcription error: {str(e)}")
        raise
        
async def synthesize_speech(text: str, language: str) -> bytes:
    #Synthesizes speech from text and returns the audio data as bytes.
    try:
        speech_config = speechsdk.SpeechConfig(subscription=TTS_SPEECH_KEY, region=TTS_SPEECH_REGION)
        speech_config.set_speech_synthesis_output_format(speechsdk.SpeechSynthesisOutputFormat.Audio16Khz32KBitRateMonoMp3)
        synthesizer = speechsdk.SpeechSynthesizer(speech_config=speech_config, audio_config=None)

        voice_map = {
            "en": "en-US-JennyNeural",
            "am": "am-ET-MekdesNeural"
        }
        voice = voice_map.get(language, "en-US-JennyNeural")
        
        ssml_string = (
            f"<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='{language}'>"
            f"<voice name='{voice}'>{text}</voice></speak>"
        )

        result = await asyncio.get_event_loop().run_in_executor(None, lambda: synthesizer.speak_ssml_async(ssml_string).get())

        if result.reason == speechsdk.ResultReason.SynthesizingAudioCompleted:
            return result.audio_data
        elif result.reason == speechsdk.ResultReason.Canceled:
            cancellation_details = result.cancellation_details
            logger.error(f"Speech synthesis canceled: {cancellation_details.reason}")
            if cancellation_details.reason == speechsdk.CancellationReason.Error:
                logger.error(f"Error details: {cancellation_details.error_details}")
            raise Exception(f"Speech synthesis canceled: {cancellation_details.reason}")
    except Exception as e:
        logger.error(f"Speech synthesis error: {str(e)}")
        raise

# --- Blob Utils ---
async def get_blob_service_client() -> BlobServiceClient:
    #Initializes and returns an async BlobServiceClient.
    return BlobServiceClient.from_connection_string(STORAGE_CONN_STRING)

def generate_sas_url(container_name: str, blob_name: str, hours: int = 24) -> str:
    #Generates a Shared Access Signature (SAS) URL for a blob.
    try:
        sas_token = generate_blob_sas(
            account_name=STORAGE_ACCOUNT,
            container_name=container_name,
            blob_name=blob_name,
            account_key=STORAGE_ACCOUNT_KEY,
            permission=BlobSasPermissions(read=True),
            expiry=datetime.utcnow() + timedelta(hours=hours)
        )
        return f"https://{STORAGE_ACCOUNT}.blob.core.windows.net/{container_name}/{blob_name}?{sas_token}"
    except Exception as e:
        logger.error(f"SAS URL generation error: {str(e)}")
        raise

# --- Translation Utils ---
async def translate_text(text: str, target_lang: str) -> str:
    #Translates text using Azure Translator service.
    try:
        url = TRANSLATOR_ENDPOINT + "/translate"
        params = {
            'api-version': '3.0',
            'to': target_lang
        }
        headers = {
            'Ocp-Apim-Subscription-Key': TRANSLATOR_KEY,
            'Ocp-Apim-Subscription-Region': TRANSLATOR_REGION,
            'Content-type': 'application/json',
            'X-ClientTraceId': str(uuid.uuid4())
        }
        body = [{'text': text}]
        
        async with aiohttp.ClientSession() as session:
            async with session.post(url, headers=headers, json=body, params=params) as response:
                response.raise_for_status()
                translation_result = await response.json()
                
        if translation_result and len(translation_result[0]['translations']) > 0:
            return translation_result[0]['translations'][0]['text']
        else:
            return "Translation failed."
    except Exception as e:
        logger.error(f"Translation error: {str(e)}")
        raise
