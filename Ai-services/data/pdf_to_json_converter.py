import os
import json
import re
from pathlib import Path
from PyPDF2 import PdfReader

def clean_value(value_str: str):
    """
    Performs aggressive cleaning on a string value extracted from the PDF.
    """
    # Remove all newlines and multiple spaces, then strip leading/trailing spaces.
    cleaned_value = re.sub(r'\s+', ' ', value_str).strip()
    
    # Remove any extra spaces around quotation marks
    cleaned_value = re.sub(r'(\S)\s+"', r'\1"', cleaned_value)
    cleaned_value = re.sub(r'"\s+(\S)', r'"\1', cleaned_value)

    # Replace specific non-standard characters with a standard space
    cleaned_value = cleaned_value.replace('\n', ' ').replace('\r', ' ')
    
    return cleaned_value

def parse_and_fix_json_from_pdf(pdf_path: Path, output_dir: Path):
    """
    Extracts text from a PDF, finds key-value pairs, and reconstructs a valid JSON.
    """
    print(f"\n📄 Processing file: '{pdf_path.name}'")
    
    try:
        reader = PdfReader(pdf_path)
        text = ""
        for page in reader.pages:
            text += page.extract_text() or ""
        
        # Aggressively normalize the entire text to a single line.
        cleaned_text = re.sub(r'\s+', ' ', text).strip()

        # Regex to find all objects enclosed in { ... }
        # This pattern is more flexible and can handle nested structures.
        potential_json_blocks = re.findall(r'(\s*\{[^{}]*?(?:\{[^{}]*\}|\[.*?\])?[^{}]*\})', cleaned_text, re.DOTALL)
        
        reconstructed_objects = []

        for block in potential_json_blocks:
            reconstructed_dict = {}
            
            # Find key-value pairs in the block
            # This regex looks for a quoted key, a colon, and then a value.
            # It's designed to be robust and capture values that might contain
            # newlines or special characters, which we handle later.
            kv_pairs = re.findall(r'"([^"]+)"\s*:\s*(".*?(?<!\\)"|\[.*?\]|\{[^{}]*?\}|[^,{}]+)', block, re.DOTALL)
            
            for key, value_str in kv_pairs:
                try:
                    # Clean the value string first to remove bad characters
                    cleaned_val_str = clean_value(value_str)
                    
                    # Then try to load it as valid JSON
                    value = json.loads(cleaned_val_str)
                except (json.JSONDecodeError, ValueError):
                    # If it's not a valid JSON literal (e.g., a simple string),
                    # remove any extra quotes and use the cleaned string as the value.
                    value = cleaned_val_str.strip('"').strip()
                
                reconstructed_dict[key] = value

            if reconstructed_dict:
                reconstructed_objects.append(reconstructed_dict)

        if reconstructed_objects:
            output_file_name = pdf_path.stem + ".json"
            output_file_path = output_dir / output_file_name
            with open(output_file_path, 'w', encoding='utf-8') as f:
                json.dump(reconstructed_objects, f, indent=4, ensure_ascii=False)
            print(f"✅ Successfully processed '{pdf_path.name}' -> '{output_file_name}'.")
        else:
            print(f"❌ Failed to find or fix JSON in '{pdf_path.name}'.")

    except Exception as e:
        print(f"❌ An unexpected error occurred while processing '{pdf_path.name}': {e}")

def main():
    script_dir = Path(__file__).parent
    pdf_input_path = script_dir
    output_json_path = script_dir

    print(f"Processing PDF files in: '{pdf_input_path}'")

    pdf_files = list(pdf_input_path.glob("*.pdf"))
    if not pdf_files:
        print(f"No PDF files found in '{pdf_input_path}'.")
        return

    print(f"\nFound {len(pdf_files)} PDF files. Starting processing...\n")

    for pdf_file in pdf_files:
        parse_and_fix_json_from_pdf(pdf_file, output_json_path)

    print("\nProcessing complete. The JSON files are located in the same directory.")

if __name__ == "__main__":
    main()