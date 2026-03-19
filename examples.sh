#!/bin/bash

# Example usage of the agent orchestrator

echo "Agent Orchestrator - Examples"
echo "=============================="
echo ""

# Make sure API key is set
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "Error: ANTHROPIC_API_KEY not set"
    echo "Run: export ANTHROPIC_API_KEY='your-key'"
    exit 1
fi

echo "1. Simple task (Fast mode expected):"
echo "   ./opensack \"Add a health check endpoint at /health\""
echo ""

echo "2. Medium complexity (Fast mode expected):"
echo "   ./opensack \"Add exponential backoff retries to the HTTP client\""
echo ""

echo "3. Complex task (Consensus mode expected):"
echo "   ./opensack \"Design a microservices architecture for a real-time chat application with authentication, message persistence, and file sharing\""
echo ""

echo "4. Architectural decision (Consensus mode expected):"
echo "   ./opensack \"Evaluate different approaches for implementing real-time notifications. Compare WebSockets, Server-Sent Events, and polling. Consider tradeoffs.\""
echo ""

echo "5. Refactoring (Consensus mode expected):"
echo "   ./opensack \"Refactor the entire authentication system to support OAuth2, SAML, and custom providers while maintaining backward compatibility\""
echo ""

echo "Run any of these commands to test the orchestrator!"
