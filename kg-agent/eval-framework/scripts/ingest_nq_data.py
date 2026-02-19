import json
import subprocess
import os
import tempfile

def ingest_chunks(json_file):
    """
    Ingest all chunks using existing CLI.
    
    Args:
        json_file: Path to prepared JSON file
    """
    with open(json_file, 'r') as f:
        chunks = json.load(f)
    
    total = len(chunks)
    
    for i, chunk in enumerate(chunks):
        # Create temporary file with content
        with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as tmp:
            tmp.write(chunk['content'])
            tmp_path = tmp.name
        
        try:
            customMetadata = {
                "question": chunk['metadata']['question'],
                "shortAnswer": chunk['metadata']['short_answer']
            }
            # Call existing ingestion CLI
            result = subprocess.run([
                'go', 'run', 'cmd/ingest/main.go',
                '-insert-doc',
                '-filePath', tmp_path,
                '-chunkSize', '20000',
                '-chunkOverlap', '0',
                '-customMetadata', json.dumps(customMetadata)
            ], capture_output=True, text=True)
            
            if result.returncode == 0:
                if (i + 1) % 100 == 0:
                    print(f"Ingested {i + 1}/{total} chunks...")
            else:
                print(f"Error ingesting chunk {chunk['chunk_id']}: {result.stderr}")
        
        finally:
            # Clean up temp file
            os.unlink(tmp_path)
    
    print(f"Ingestion complete: {total} chunks")

if __name__ == "__main__":
    ingest_chunks('eval-framework/.data/natural_questions_prepared.json')
