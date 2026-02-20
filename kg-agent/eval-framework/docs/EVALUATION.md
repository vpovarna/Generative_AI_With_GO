# Search Evaluation Specification

## Overview

This document outlines the evaluation plan for testing the KG Agent's search quality using the Natural Questions dataset. The evaluation focuses on **information retrieval performance** (search accuracy) rather than end-to-end answer quality.

**Dataset:** Natural Questions (Filtered)
- **Size:** 86,212 questions with long/short answers
- **Source:** All questions from Wikipedia articles
- **Structure:** Question + Long Answer + Short Answer

**Goal:** Measure if the hybrid search system can accurately retrieve the correct answer chunk for a given question.

---

## Data Structure

### Input Format (Natural Questions CSV)

```csv
question,long_answer,short_answer
"which is the most common use of opt-in e-mail marketing","A common example of permission marketing is a newsletter sent to an advertising firm's customers. Such newsletters inform customers of upcoming events or promotions, or new products...","A newsletter sent to an advertising firm's customers"
```

### Prepared JSON Format

Transform each row into:

```json
[
  {
    "chunk_id": "nq_000000",
    "document_id": "natural_questions",
    "content": "`` Don't You (Forget About Me) '' is a 1985 pop song performed by Scottish rock band Simple Minds. The song is best known for being played during the opening and closing credits of the John Hughes film The Breakfast Club. It was written and composed by producer Keith Dorsey and Steve Schiff, the latter of whom was a guitarist and songwriter from the Nina Hagen band.",
    "metadata": {
      "question": "what film has the song don't you forget about me",
      "short_answer": "The Breakfast Club"
    }
  }
]
```

**Field Descriptions:**
- `chunk_id`: Unique identifier for the chunk (format: `nq_XXXXXX`)
- `document_id`: Source document (all use `"natural_questions"`)
- `content`: The long answer text (what gets embedded and searched)
- `metadata.question`: Original question (for debugging and reference)
- `metadata.short_answer`: Short answer (for future LLM judge evaluation)

---

### Evaluation

**Script:** `scripts/eval_search.py`

```python
import requests
import json
from typing import List, Dict
from collections import defaultdict

class SearchEvaluator:
    def __init__(self, search_api_url: str, test_data_path: str, sample_size: int = 1000):
        self.search_api_url = search_api_url
        self.sample_size = sample_size
        
        # Load test data
        with open(test_data_path, 'r') as f:
            all_data = json.load(f)
            self.test_data = all_data[:sample_size]
        
        print(f"Loaded {len(self.test_data)} test cases")
    
    def evaluate_search(self, search_type: str, limit: int = 5) -> List[Dict]:
        """
        Evaluate a specific search type.
        
        Args:
            search_type: 'semantic', 'keyword', or 'hybrid'
            limit: Number of results to retrieve
        
        Returns:
            List of result dictionaries
        """
        results = []
        
        for i, qa in enumerate(self.test_data):
            question = qa['metadata']['question']
            expected_chunk_id = qa['chunk_id']
            
            # Call search API
            try:
                response = requests.post(
                    f'{self.search_api_url}/{search_type}',
                    json={'query': question, 'limit': limit},
                    timeout=10
                )
                response.raise_for_status()
                search_results = response.json()['result']
            except Exception as e:
                print(f"Error on question {i}: {e}")
                search_results = []
            
            # Find if expected chunk is in results
            found = False
            rank = -1
            returned_ids = []
            
            for j, result in enumerate(search_results):
                returned_ids.append(result.get('chunk_id', 'unknown'))
                if result.get('chunk_id') == expected_chunk_id:
                    found = True
                    rank = j + 1
                    break
            
            results.append({
                'question_id': qa['metadata']['question_id'],
                'question': question,
                'expected_chunk_id': expected_chunk_id,
                'found': found,
                'rank': rank,
                'returned_chunk_ids': returned_ids
            })
            
            if (i + 1) % 100 == 0:
                print(f"  Evaluated {i + 1}/{len(self.test_data)} questions...")
        
        return results
    
    def calculate_metrics(self, results: List[Dict]) -> Dict:
        """
        Calculate evaluation metrics.
        
        Returns:
            Dictionary with recall@1, recall@5, and MRR
        """
        total = len(results)
        
        recall_at_1 = sum(1 for r in results if r['rank'] == 1) / total
        recall_at_5 = sum(1 for r in results if r['found']) / total
        
        # Mean Reciprocal Rank
        mrr = sum(1/r['rank'] for r in results if r['rank'] > 0) / total
        
        # Additional stats
        not_found = sum(1 for r in results if not r['found'])
        avg_rank = sum(r['rank'] for r in results if r['rank'] > 0) / max(1, (total - not_found))
        
        return {
            'total_questions': total,
            'recall_at_1': recall_at_1,
            'recall_at_5': recall_at_5,
            'mrr': mrr,
            'not_found': not_found,
            'not_found_pct': not_found / total,
            'avg_rank_when_found': avg_rank
        }
    
    def run_full_evaluation(self):
        """
        Run evaluation for all search types and print comparison.
        """
        search_types = ['semantic', 'keyword', 'hybrid']
        all_results = {}
        
        for search_type in search_types:
            print(f"\n{'='*60}")
            print(f"Evaluating {search_type.upper()} search...")
            print(f"{'='*60}")
            
            results = self.evaluate_search(search_type, limit=5)
            metrics = self.calculate_metrics(results)
            
            all_results[search_type] = {
                'results': results,
                'metrics': metrics
            }
            
            # Print metrics
            print(f"\n📊 {search_type.upper()} SEARCH METRICS:")
            print(f"  Recall@1:  {metrics['recall_at_1']:.2%}")
            print(f"  Recall@5:  {metrics['recall_at_5']:.2%}")
            print(f"  MRR:       {metrics['mrr']:.3f}")
            print(f"  Not Found: {metrics['not_found']} ({metrics['not_found_pct']:.1%})")
            print(f"  Avg Rank:  {metrics['avg_rank_when_found']:.2f}")
        
        # Print comparison
        print(f"\n{'='*60}")
        print("📈 COMPARISON")
        print(f"{'='*60}")
        print(f"{'Metric':<20} {'Semantic':>12} {'Keyword':>12} {'Hybrid':>12}")
        print("-" * 60)
        
        for metric in ['recall_at_1', 'recall_at_5', 'mrr']:
            values = [all_results[st]['metrics'][metric] for st in search_types]
            if metric == 'mrr':
                print(f"{metric.upper():<20} {values[0]:>12.3f} {values[1]:>12.3f} {values[2]:>12.3f}")
            else:
                print(f"{metric.upper():<20} {values[0]:>11.1%} {values[1]:>11.1%} {values[2]:>11.1%}")
        
        # Save results to file
        with open('evaluation_results.json', 'w') as f:
            json.dump(all_results, f, indent=2)
        
        print(f"\n✅ Results saved to: evaluation_results.json")
        
        return all_results

if __name__ == "__main__":
    evaluator = SearchEvaluator(
        search_api_url='http://localhost:8082/search/v1',
        test_data_path='data/natural_questions_prepared.json',
        sample_size=1000  # Test on 1000 questions
    )
    
    evaluator.run_full_evaluation()
```

**Execution:**
```bash
cd kg-agent
python scripts/eval_search.py
```

---

## Evaluation Metrics

### Primary Metrics

#### 1. **Recall@1** (Top-1 Accuracy)
```
Recall@1 = (# questions where correct chunk is ranked #1) / (total questions)
```

**Interpretation:**
- **>80%**: Excellent - Most queries get the right answer immediately
- **70-80%**: Good - Acceptable for most use cases
- **<70%**: Needs improvement - Too many wrong results at the top

**Target:** >75%

#### 2. **Recall@5** (Top-5 Coverage)
```
Recall@5 = (# questions where correct chunk is in top 5) / (total questions)
```

**Interpretation:**
- **>95%**: Excellent - Nearly all relevant chunks are retrieved
- **90-95%**: Good - Most queries succeed
- **85-90%**: Acceptable - Room for improvement
- **<85%**: Poor - Search needs tuning

**Target:** >90%

#### 3. **Mean Reciprocal Rank (MRR)**
```
MRR = Average of (1 / rank) for all questions

Example:
  Q1: Correct at rank 1 → 1/1 = 1.000
  Q2: Correct at rank 3 → 1/3 = 0.333
  Q3: Not found (rank ∞) → 0.000
  MRR = (1.000 + 0.333 + 0.000) / 3 = 0.444
```

**Interpretation:**
- **>0.85**: Excellent - Correct answers are typically #1 or #2
- **0.75-0.85**: Good - Decent ranking quality
- **0.60-0.75**: Acceptable - Correct answers often in positions 2-3
- **<0.60**: Poor - Ranking needs improvement

**Target:** >0.75

### Secondary Metrics

#### 4. **Not Found Rate**
```
Not Found = (# questions where correct chunk not in top 5) / (total questions)
```

**Target:** <10%

#### 5. **Average Rank (When Found)**
```
Avg Rank = Average position of correct chunk (for queries where it's found)
```

**Target:** <2.5

---

## Expected Results

### Baseline Performance Targets

| Search Type | Recall@1 | Recall@5 | MRR | Notes |
|-------------|----------|----------|-----|-------|
| **Semantic** | 65-75% | 85-90% | 0.70-0.75 | Vector embeddings capture meaning |
| **Keyword** | 50-65% | 75-85% | 0.60-0.70 | Works for exact term matches |
| **Hybrid** | **75-85%** | **90-95%** | **0.75-0.85** | Best of both worlds |

**Key Insight:** Hybrid search should outperform individual methods by 5-10% on all metrics.

### Success Criteria

✅ **Minimum Acceptable:**
- Hybrid Recall@5: >90%
- Hybrid MRR: >0.75

✅ **Good Performance:**
- Hybrid Recall@1: >75%
- Hybrid Recall@5: >92%
- Hybrid MRR: >0.80

✅ **Excellent Performance:**
- Hybrid Recall@1: >80%
- Hybrid Recall@5: >95%
- Hybrid MRR: >0.85

---

## Evaluation Output

### Console Output Example

```
============================================================
Evaluating HYBRID search...
============================================================
  Evaluated 100/1000 questions...
  Evaluated 200/1000 questions...
  ...
  Evaluated 1000/1000 questions...

📊 HYBRID SEARCH METRICS:
  Recall@1:  78.3%
  Recall@5:  93.1%
  MRR:       0.823
  Not Found: 69 (6.9%)
  Avg Rank:  2.12

============================================================
📈 COMPARISON
============================================================
Metric               Semantic     Keyword      Hybrid
------------------------------------------------------------
RECALL_AT_1            72.5%        58.3%        78.3%
RECALL_AT_5            88.2%        79.4%        93.1%
MRR                    0.762        0.684        0.823

✅ Results saved to: evaluation_results.json
```

### JSON Output Structure

**File:** `evaluation_results.json`

```json
{
  "semantic": {
    "metrics": {
      "total_questions": 1000,
      "recall_at_1": 0.725,
      "recall_at_5": 0.882,
      "mrr": 0.762,
      "not_found": 118,
      "not_found_pct": 0.118,
      "avg_rank_when_found": 2.34
    },
    "results": [
      {
        "question_id": "1",
        "question": "which is the most common use of opt-in e-mail marketing",
        "expected_chunk_id": "nq_000001",
        "found": true,
        "rank": 1,
        "returned_chunk_ids": ["nq_000001", "nq_004523", "nq_002341", ...]
      }
    ]
  },
  "keyword": {...},
  "hybrid": {...}
}
```

---

## Troubleshooting

### Low Recall@5 (<85%)

**Possible causes:**
1. **Embedding quality issues**
   - Check: Are embeddings being generated correctly?
   - Solution: Verify Bedrock Titan model configuration

2. **Search query not matching content**
   - Check: Compare question text vs long_answer text
   - Solution: Ensure proper text preprocessing

3. **Index corruption**
   - Check: Are all chunks indexed in PostgreSQL?
   - Solution: Re-run ingestion

### Low MRR (<0.70) but High Recall@5

**Diagnosis:** Correct chunks are found but ranked poorly.

**Possible causes:**
1. **RRF weights not optimized**
   - Solution: Tune RRF k-parameter or semantic/keyword weights

2. **Distance calculation issues**
   - Check: Are cosine distances properly normalized?
   - Solution: Verify `DistanceToScore` function

### Semantic << Keyword Performance

**Diagnosis:** Embeddings not capturing semantic meaning.

**Possible causes:**
1. **Wrong embedding model**
   - Solution: Verify using correct Titan model

2. **Text truncation**
   - Check: Are long answers being truncated?
   - Solution: Ensure full content is embedded

---

## Future Enhancements

### Phase 4: LLM as Judge (Optional)

For end-to-end answer quality evaluation:

**Script:** `scripts/eval_llm_judge.py`

```python
# Sample 100 questions
# Call full agent API (not just search)
# Compare agent response with short_answer using LLM judge
# Score on 1-5 scale

Target: Average score >4.0
```

### Phase 5: Error Analysis

**Script:** `scripts/analyze_errors.py`

```python
# Find all questions where Recall@5 = False
# Categorize failures:
#   - Missing chunks (ingestion issue)
#   - Wrong embeddings (semantic issue)
#   - Poor query (question ambiguity)
```

### Phase 6: A/B Testing

Test improvements:
- Different embedding models
- Different RRF weights
- Different reranking strategies

Compare metrics before/after changes.

---

## Maintenance

### Regular Evaluation

Run evaluation:
- **Weekly** during active development
- **After** any search-related code changes
- **Before** production deployments

### Regression Testing

Maintain a **golden dataset** (100-200 questions) for quick smoke tests:

```bash
python scripts/eval_search.py --sample-size 100 --quick-test
```

Target: Should complete in <2 minutes.

---

## References

- Natural Questions Dataset: https://ai.google.com/research/NaturalQuestions
- Mean Reciprocal Rank: https://en.wikipedia.org/wiki/Mean_reciprocal_rank
- Information Retrieval Metrics: https://www.evidentlyai.com/ranking-metrics
- RRF (Reciprocal Rank Fusion): https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf
