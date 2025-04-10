#!/bin/bash

# AGENTIC AI SETUP FOR GITHUB CODESPACE (Ubuntu Linux)
# ----------------------------------------------------
# This script installs all required components to enable app generation,
# auto-deployment using GitHub Actions, and Kubernetes deployment via AWS.
#
# PROMPT-BASED STEPS WILL ASK FOR APPROVAL BEFORE CONTINUING

set -e

# Function: prompt for user approval with default yes
confirm() {
    echo -n "$1 (Y/n): "
    read -r response
    response=${response:-y}
    case $response in
        [yY]) true ;;
        *) echo "Skipped by user." && return 1 ;;
    esac
}

# 1. Install base CLI tools
echo "\nStep 1: Installing git, gh, awscli, and prerequisites for kubectl"
confirm "Proceed with installing base CLI tools" && \
  sudo apt update && sudo apt install -y git gh awscli apt-transport-https ca-certificates curl

# Install kubectl if not already installed
if ! command -v kubectl &> /dev/null; then
  echo "\nInstalling kubectl from Kubernetes repo"
  sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
  echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | \
    sudo tee /etc/apt/sources.list.d/kubernetes.list
  sudo apt update && sudo apt install -y kubectl
else
  echo "✅ kubectl is already installed, skipping."
fi

# 2. Install Ollama (Local LLM backend)
echo "\nStep 2: Installing Ollama for running local LLMs"
if ! command -v ollama &> /dev/null; then
  confirm "Ollama not found. Install it now?" && curl -fsSL https://ollama.com/install.sh | sh
else
  echo "✅ Ollama is already installed."
fi

# Check if Ollama is running and start if not
if ! pgrep -x "ollama" > /dev/null; then
  echo "Starting Ollama server..."
  nohup ollama serve > /tmp/ollama.log 2>&1 &
  sleep 2
else
  echo "✅ Ollama is already running."
fi

# 3. Run codellama model
echo "\nStep 3: Pulling CodeLlama model"
if ! ollama list | grep -q codellama; then
  confirm "Pull and run CodeLlama model (400MB+)?" && ollama run codellama &
else
  echo "✅ CodeLlama already pulled."
fi

# 4. Install Auto-GPT
echo "\nStep 4: Cloning Auto-GPT and configuring it"
if [ ! -d "Auto-GPT" ]; then
  confirm "Clone and set up Auto-GPT?" && \
    git clone https://github.com/Torantulino/Auto-GPT.git && \
    cd Auto-GPT && \
    python3 -m venv venv && \
    source venv/bin/activate && \
    pip install -r requirements.txt && \
    cat << EOF > .env
OPENAI_API_BASE=http://localhost:11434/v1
OPENAI_API_KEY=ollama
EOF
else
  echo "✅ Auto-GPT directory already exists, skipping clone."
  cd Auto-GPT
fi

# 5. Configure AWS for Kubeconfig and Deployment
echo "\nStep 5: AWS credentials setup"
confirm "Configure AWS CLI with access key/secret (will prompt you)" && aws configure

# 6. Sample GitHub Action YAML file
echo "\nStep 6: Writing sample GitHub Actions deploy file"
mkdir -p ../.github/workflows
cat << YAML > ../.github/workflows/deploy.yml
name: Deploy to Kubernetes

on:
  push:
    branches:
      - main

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v2
      with:
        aws-access-key-id: \${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: \${{ secrets.AWS_SECRET_ACCESS_KEY }}
        region: us-east-1
    - name: Setup Kubeconfig
      run: aws eks update-kubeconfig --name your-cluster-name
    - name: Deploy to Kubernetes
      run: kubectl apply -f k8s/
YAML

# 7. Complete message
echo "\n✅ Agentic AI environment setup complete."
echo "- Auto-GPT is ready with local LLM via Ollama"
echo "- AWS CLI and kubectl installed"
echo "- GitHub Actions ready to deploy via k8s"
echo "\n👉 Next Step: Customize 'Auto-GPT' goals to read your repo and write app code."
