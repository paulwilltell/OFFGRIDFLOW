"""
OffGridFlow AI Evaluation Framework

This script evaluates AI responses from the carbon emissions platform using:
- Relevance: How well responses address carbon accounting queries
- Coherence: Logical structure and clarity of responses  
- Groundedness: Factual accuracy based on emission data
"""

import os
import json
from azure.ai.evaluation import (
    evaluate,
    RelevanceEvaluator,
    CoherenceEvaluator,
    GroundednessEvaluator,
    OpenAIModelConfiguration
)

def load_dataset(queries_path: str, responses_path: str, output_path: str) -> str:
    """
    Combine queries and responses into evaluation dataset format.
    
    Args:
        queries_path: Path to test queries JSONL file
        responses_path: Path to test responses JSONL file
        output_path: Path to save combined dataset
        
    Returns:
        Path to combined dataset
    """
    queries = []
    responses = []
    
    with open(queries_path, 'r', encoding='utf-8') as f:
        for line in f:
            queries.append(json.loads(line))
    
    with open(responses_path, 'r', encoding='utf-8') as f:
        for line in f:
            responses.append(json.loads(line))
    
    # Combine into evaluation format
    combined = []
    for q, r in zip(queries, responses):
        combined.append({
            "query": q["query"],
            "response": r["response"],
            "context": r["response"]  # Using response as context for groundedness
        })
    
    with open(output_path, 'w', encoding='utf-8') as f:
        for item in combined:
            f.write(json.dumps(item) + '\n')
    
    return output_path

def main():
    """Run comprehensive evaluation of OffGridFlow AI responses."""
    
    # Configure OpenAI model for evaluation (uses environment variable)
    # Set OPENAI_API_KEY environment variable before running
    model_config = OpenAIModelConfiguration(
        type="openai",
        model=os.getenv("OPENAI_MODEL", "gpt-4o-mini"),  # Default to gpt-4o-mini
        api_key=os.environ.get("OPENAI_API_KEY")
    )
    
    # Prepare dataset
    print("📊 Preparing evaluation dataset...")
    dataset_path = load_dataset(
        queries_path="test_queries.jsonl",
        responses_path="test_responses.jsonl",
        output_path="evaluation_dataset.jsonl"
    )
    print(f"✅ Dataset ready: {dataset_path}")
    
    # Initialize evaluators
    print("\n🔧 Initializing evaluators...")
    relevance_eval = RelevanceEvaluator(model_config=model_config)
    coherence_eval = CoherenceEvaluator(model_config=model_config)
    groundedness_eval = GroundednessEvaluator(model_config=model_config)
    print("✅ Evaluators initialized")
    
    # Run evaluation
    print("\n🚀 Running evaluation...")
    print("   This may take a few minutes...\n")
    
    result = evaluate(
        data=dataset_path,
        evaluators={
            "relevance": relevance_eval,
            "coherence": coherence_eval,
            "groundedness": groundedness_eval
        },
        evaluator_config={
            "relevance": {
                "column_mapping": {
                    "query": "${data.query}",
                    "response": "${data.response}"
                }
            },
            "coherence": {
                "column_mapping": {
                    "query": "${data.query}",
                    "response": "${data.response}"
                }
            },
            "groundedness": {
                "column_mapping": {
                    "response": "${data.response}",
                    "context": "${data.context}"
                }
            }
        },
        output_path="evaluation_results"
    )
    
    # Display results
    print("\n" + "="*60)
    print("📊 EVALUATION RESULTS")
    print("="*60)
    
    if "metrics" in result:
        metrics = result["metrics"]
        print(f"\n🎯 Overall Metrics:")
        print(f"   Relevance Score:    {metrics.get('relevance.relevance', 'N/A')}")
        print(f"   Coherence Score:    {metrics.get('coherence.coherence', 'N/A')}")
        print(f"   Groundedness Score: {metrics.get('groundedness.groundedness', 'N/A')}")
    
    print(f"\n💾 Detailed results saved to: evaluation_results/")
    print(f"   - evaluation_results.jsonl (row-level scores)")
    print(f"   - evaluation_results_metrics.json (aggregate metrics)")
    
    print("\n✅ Evaluation complete!")
    print("="*60)
    
    return result

if __name__ == "__main__":
    # Ensure we're in the evaluation directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(script_dir)
    
    # Check for API key
    if not os.environ.get("OPENAI_API_KEY"):
        print("⚠️  Warning: OPENAI_API_KEY environment variable not set")
        print("   Set it with: $env:OPENAI_API_KEY='your-key-here' (PowerShell)")
        print("   Or use Azure OpenAI by modifying model_config in the script")
        exit(1)
    
    main()
