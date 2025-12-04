# Quick Start: Evaluation & Tracing

This guide helps you get started with OffGridFlow's evaluation framework and distributed tracing.

## 🎯 Evaluation Framework

### Setup (5 minutes)

1. **Navigate to evaluation directory:**
   ```powershell
   cd evaluation
   ```

2. **Install Python dependencies:**
   ```powershell
   pip install -r requirements.txt
   ```

3. **Set your OpenAI API key:**
   ```powershell
   $env:OPENAI_API_KEY = "sk-..."
   ```

### Run Evaluation

```powershell
python run_evaluation.py
```

**What it evaluates:**
- ✅ Relevance: How well AI responses address carbon accounting queries
- ✅ Coherence: Logical structure and clarity of responses
- ✅ Groundedness: Factual accuracy based on emission data

**Results:**
- `evaluation_results/evaluation_results.jsonl` - Row-level scores
- `evaluation_results/evaluation_results_metrics.json` - Aggregate metrics

### Customize Test Data

Edit these files to add your own test cases:
- `test_queries.jsonl` - Add your test questions
- `test_responses.jsonl` - Add corresponding AI responses

Format:
```json
{"query": "Your test question"}
{"query": "Your test question", "response": "AI response to evaluate"}
```

---

## 🔍 Distributed Tracing

### Setup (2 minutes)

**1. Open AI Toolkit Trace Viewer:**

In VS Code:
- Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on Mac)
- Type: "AI Toolkit: Open Trace Viewer"
- Press Enter

This starts the trace collector at `http://localhost:4318`

**2. Run your application:**

```powershell
go run ./cmd/api
```

Look for this in the logs:
```
[offgridflow] tracing enabled (endpoint: http://localhost:4318)
```

**3. View traces:**

The AI Toolkit trace viewer will show:
- Request timelines
- Operation hierarchy
- Performance metrics
- Error tracking

### Configuration

**Disable tracing:**
```powershell
$env:OFFGRIDFLOW_TRACING_ENABLED = "false"
```

**Custom endpoint:**
```powershell
$env:OTEL_EXPORTER_OTLP_ENDPOINT = "http://custom-host:4318"
```

---

## 📚 Learn More

- **Evaluation**: See `evaluation/README.md` for detailed configuration
- **Tracing**: See `docs/TRACING.md` for advanced usage and best practices

## 🆘 Troubleshooting

### Evaluation Issues

**Problem:** `ModuleNotFoundError`
```powershell
pip install -r evaluation/requirements.txt
```

**Problem:** `OPENAI_API_KEY not set`
```powershell
$env:OPENAI_API_KEY = "your-key-here"
```

### Tracing Issues

**Problem:** No traces appearing
1. Ensure AI Toolkit trace viewer is open
2. Check trace collector is running on port 4318
3. Verify `OFFGRIDFLOW_TRACING_ENABLED` is not set to "false"

**Problem:** Connection refused
- Restart VS Code
- Reopen AI Toolkit trace viewer

---

## 🎓 Next Steps

1. **Run evaluation** on your actual AI responses
2. **Add custom evaluators** for business-specific metrics
3. **Instrument your code** with additional trace spans
4. **Monitor production** with trace sampling

Happy evaluating and tracing! 🚀
