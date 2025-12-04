# OffGridFlow Evaluation Framework

This directory contains the evaluation framework for OffGridFlow's AI-powered carbon emissions platform.

## Overview

The evaluation framework assesses AI response quality across three key dimensions:

- **Relevance**: How well responses address carbon accounting queries
- **Coherence**: Logical structure and clarity of responses
- **Groundedness**: Factual accuracy based on emission data

## Files

- `test_queries.jsonl` - Test queries covering emissions calculations, compliance, and carbon accounting
- `test_responses.jsonl` - Sample AI responses for evaluation
- `run_evaluation.py` - Main evaluation script
- `requirements.txt` - Python dependencies

## Setup

1. **Install dependencies:**
   ```powershell
   pip install -r requirements.txt
   ```

2. **Set your OpenAI API key:**
   ```powershell
   $env:OPENAI_API_KEY = "your-api-key-here"
   ```

   Or to use Azure OpenAI, modify `model_config` in `run_evaluation.py`:
   ```python
   from azure.ai.evaluation import AzureOpenAIModelConfiguration
   
   model_config = AzureOpenAIModelConfiguration(
       azure_deployment="your-deployment-name",
       azure_endpoint="https://your-resource.openai.azure.com/",
       api_key=os.environ["AZURE_OPENAI_API_KEY"],
       api_version="2025-04-01-preview"
   )
   ```

## Running Evaluation

```powershell
cd evaluation
python run_evaluation.py
```

## Results

Evaluation results are saved to `evaluation_results/`:
- `evaluation_results.jsonl` - Row-level scores for each query/response pair
- `evaluation_results_metrics.json` - Aggregate metrics and statistics

## Customizing Evaluation

### Add Custom Evaluators

Create a custom code-based evaluator:

```python
class CustomEvaluator:
    def __init__(self):
        pass
    
    def __call__(self, *, response: str, **kwargs):
        # Your evaluation logic here
        return {"custom_metric": your_score}
```

Add to the `evaluate()` call:

```python
result = evaluate(
    data=dataset_path,
    evaluators={
        "relevance": relevance_eval,
        "coherence": coherence_eval,
        "groundedness": groundedness_eval,
        "custom": CustomEvaluator()
    },
    evaluator_config={
        # ... existing config ...
        "custom": {
            "column_mapping": {
                "response": "${data.response}"
            }
        }
    },
    output_path="evaluation_results"
)
```

### Modify Test Dataset

Edit `test_queries.jsonl` and `test_responses.jsonl` to add your own test cases:

```json
{"query": "Your test question here"}
```

```json
{"query": "Your test question here", "response": "Expected AI response here"}
```

## Metrics Explanation

### Relevance (1-5 scale)
- 5: Perfectly addresses the query
- 4: Mostly relevant with minor gaps
- 3: Partially relevant
- 2: Minimally relevant
- 1: Not relevant

### Coherence (1-5 scale)
- 5: Perfectly logical and clear
- 4: Mostly coherent with minor issues
- 3: Somewhat coherent
- 2: Poorly structured
- 1: Incoherent

### Groundedness (1-5 scale)
- 5: Fully grounded in context
- 4: Mostly grounded
- 3: Partially grounded
- 2: Minimally grounded
- 1: Not grounded
