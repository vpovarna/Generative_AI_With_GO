import json
import csv
import argparse
import random

def prepare_natural_questions(input_csv, output_json, sample_size=None, random_seed=42):
    """
    Convert Natural Questions CSV to evaluation JSON format.
    
    Args:
        input_csv: Path to Natural Questions filtered dataset
        output_json: Path to output JSON file
        sample_size: Number of questions to use (None = all)
        random_seed: Random seed for reproducible sampling (default: 42)
    """
    # Read all data
    all_rows = []
    with open(input_csv, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        all_rows = list(reader)
    
    total_available = len(all_rows)
    print(f"Total questions available: {total_available:,}")
    
    # Sample if requested
    if sample_size and sample_size < total_available:
        random.seed(random_seed)
        selected_indices = random.sample(range(total_available), k=sample_size)
        selected_indices.sort()  # Sort for deterministic chunk_ids
        selected_rows = [all_rows[i] for i in selected_indices]
        print(f"Randomly selected {sample_size:,} questions (seed: {random_seed})")
    else:
        selected_rows = all_rows
        print(f"Using all {total_available:,} questions")
    
    # Convert to chunks
    chunks = []
    for i, row in enumerate(selected_rows):
        chunks.append({
            "chunk_id": f"nq_{i:06d}",
            "document_id": "natural_questions",
            "content": row['long_answers'],
            "metadata": {
                "question": row['question'],
                "short_answer": row['short_answers']
            }
        })
    
    # Save output
    with open(output_json, 'w', encoding='utf-8') as f:
        json.dump(chunks, f, indent=2, ensure_ascii=False)
    
    print(f"Prepared {len(chunks):,} chunks")
    print(f"Output: {output_json}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description='Prepare Natural Questions dataset for evaluation'
    )
    parser.add_argument(
        '--input',
        default='eval-framework/.data/Natural-Questions-Filtered.csv',
        help='Path to input CSV file'
    )
    parser.add_argument(
        '--output',
        default='eval-framework/.data/natural_questions_prepared.json',
        help='Path to output JSON file'
    )
    parser.add_argument(
        '--sample-size',
        type=int,
        default=1000,
        help='Number of questions to prepare (default: 1000, 0 = all)'
    )
    parser.add_argument(
        '--seed',
        type=int,
        default=42,
        help='Random seed for reproducibility (default: 42)'
    )
    
    args = parser.parse_args()
    
    sample_size = None if args.sample_size == 0 else args.sample_size
    
    prepare_natural_questions(
        args.input,
        args.output,
        sample_size,
        args.seed
    )