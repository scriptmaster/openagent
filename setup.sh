#!/bin/bash

# Relaunch in bash if not already running in bash
if [ -z "$BASH_VERSION" ]; then
  exec bash "$0" "$@"
fi

set -e
CONFIG_FILE="cloud_config.env"

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

# Function: choose cloud provider for deployment
choose_cloud() {
    echo -n $'\nDeploy using (a) AWS EKS or (g) Google GKE (default): '
    read -r provider
    provider=${provider:-g}
    echo $provider
}

# 1. Install base CLI tools
if command -v git >/dev/null && command -v gh >/dev/null && command -v curl >/dev/null; then
  echo "✅ Step 1: Base CLI tools already installed. Skipping."
else
  echo $'\n🔧 Step 1: Installing git, gh, curl, and prerequisites for kubectl'
  confirm "Proceed with installing base CLI tools" && \
    sudo apt update && sudo apt install -y git gh apt-transport-https ca-certificates curl
fi

# Install kubectl if not already installed
if ! [ -x "$(which kubectl)" ]; then
  echo $'\n🔧 Installing kubectl from Kubernetes repo'
  sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
  echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | \
    sudo tee /etc/apt/sources.list.d/kubernetes.list
  sudo apt update && sudo apt install -y kubectl
else
  echo "✅ kubectl is already installed, skipping."
fi

# 2. Install Ollama (Local LLM backend)
echo $'\n🔧 Step 2: Installing Ollama for running local LLMs'
if ! [ -x "$(which ollama)" ]; then
  confirm "Ollama not found. Install it now?" && curl -fsSL https://ollama.com/install.sh | sh
else
  echo "✅ Ollama is already installed."
fi

# Check if Ollama is running and start if not
if ! curl -s http://localhost:11434 &> /dev/null; then
  echo "🚀 Starting Ollama server..."
  nohup ollama serve > /tmp/ollama.log 2>&1 &
  sleep 2
else
  echo "✅ Ollama server already running."
fi

# 3. Run or pull smaller model (mistral for Codespaces compatibility)
echo $'\n🔧 Step 3: Checking for smaller model 'mistral' (better for Codespaces)'
if ! ollama list | grep -q "mistral"; then
  echo "📥 Pulling Mistral model (lighter alternative to CodeLlama)..."
  if ! ollama pull mistral; then
    echo "❌ Failed to pull mistral model. Please check logs."
  else
    echo "✅ Mistral pulled successfully."
  fi
else
  echo "✅ Mistral model already exists."
fi

# 4. Clone Auto-GPT if needed
echo $'\n🔧 Step 4: Cloning Auto-GPT'
if [ ! -d "Auto-GPT" ]; then
  confirm "Clone Auto-GPT repo?" && git clone https://github.com/Significant-Gravitas/Auto-GPT.git
else
  echo "✅ Auto-GPT repo already exists."
fi

# 5. Set up Auto-GPT environment (Docker-based)

echo $'
🔧 Step 5: Setting up Auto-GPT (Docker-based)'

cd Auto-GPT

echo "📦 Initializing submodules..."
git submodule update --init --recursive --progress

cd autogpt_platform
if [ ! -f ".env" ]; then
  cp .env.example .env

  # Inject secure random values into the backend .env (first time only)
  sed -i "s|POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=$(openssl rand -hex 16)|" .env
  sed -i "s|JWT_SECRET=.*|JWT_SECRET=$(openssl rand -hex 32)|" .env
  sed -i "s|SECRET_KEY_BASE=.*|SECRET_KEY_BASE=$(openssl rand -hex 32)|" .env
  sed -i "s|VAULT_ENC_KEY=.*|VAULT_ENC_KEY=$(openssl rand -hex 32)|" .env
  sed -i "s|LOGFLARE_LOGGER_BACKEND_API_KEY=.*|LOGFLARE_LOGGER_BACKEND_API_KEY=$(openssl rand -hex 24)|" .env
  sed -i "s|LOGFLARE_API_KEY=.*|LOGFLARE_API_KEY=$(openssl rand -hex 24)|" .env

  # Pull Google Cloud keys from config if available
  if [ -n "$GCP_PROJECT_ID" ]; then
    sed -i "s|GOOGLE_PROJECT_ID=.*|GOOGLE_PROJECT_ID=$GCP_PROJECT_ID|" .env
  fi
  if [ -n "$GCP_PROJECT_NUMBER" ]; then
    sed -i "s|GOOGLE_PROJECT_NUMBER=.*|GOOGLE_PROJECT_NUMBER=$GCP_PROJECT_NUMBER|" .env
  fi

  echo "✅ .env created with secure autogenerated keys and config values. Please review."
else
  echo "✅ .env already exists. Skipping regeneration."
fi
echo "✅ .env created with secure autogenerated keys (POSTGRES_PASSWORD, JWT_SECRET, etc). Please review."


echo "🚀 Launching backend services with Docker Compose..."
docker compose up -d --build

cd frontend
cp .env.example .env
npm install
npm run dev

cd ../..
if [ ! -f ".env" ]; then
  cat << EOF > .env
OPENAI_API_BASE=http://localhost:11434/v1
OPENAI_API_KEY=ollama
EOF
  echo "✅ .env file created."
else
  echo "✅ .env already exists."
fi

# 6. Configure cloud provider
if [ -f "../$CONFIG_FILE" ]; then
  echo "✅ Loading saved config from $CONFIG_FILE"
  source "../$CONFIG_FILE"
else
  CLOUD=$(choose_cloud)
  echo "CLOUD=$CLOUD" > "../$CONFIG_FILE"

  if [ "$CLOUD" = "a" ]; then
    echo "\n🔧 Setting up AWS CLI for EKS..."
    sudo apt install -y awscli
    confirm "Configure AWS CLI with access key/secret (will prompt you)" && aws configure
    echo -n "Enter your EKS cluster name: " && read -r EKS_CLUSTER_NAME
    echo -n "Enter your AWS region (e.g., us-east-1): " && read -r AWS_REGION
    echo "EKS_CLUSTER_NAME=$EKS_CLUSTER_NAME" >> "../$CONFIG_FILE"
    echo "AWS_REGION=$AWS_REGION" >> "../$CONFIG_FILE"
  elif [ "$CLOUD" = "g" ]; then
    echo "\n🔧 Setting up Google Cloud CLI for GKE..."
    sudo apt install -y gnupg
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | \
      sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | \
      sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
    sudo apt update && sudo apt install -y google-cloud-sdk
    confirm "Login to Google Cloud account now (opens browser)?" && gcloud auth login
    echo -n "Enter your GCP project ID: " && read -r GCP_PROJECT_ID
    echo -n "Enter your compute zone (e.g., us-central1-a): " && read -r GCP_COMPUTE_ZONE
    echo -n "Enter your GKE cluster name: " && read -r GKE_CLUSTER_NAME
    echo "GCP_PROJECT_ID=$GCP_PROJECT_ID" >> "../$CONFIG_FILE"
    echo "GCP_COMPUTE_ZONE=$GCP_COMPUTE_ZONE" >> "../$CONFIG_FILE"
    echo "GKE_CLUSTER_NAME=$GKE_CLUSTER_NAME" >> "../$CONFIG_FILE"
  else
    echo "❌ Invalid selection. Aborting."
    exit 1
  fi
fi

# 7. Set up cluster credentials
if [ "$CLOUD" = "a" ]; then
  CLUSTER_SETUP_COMMAND="aws eks update-kubeconfig --region $AWS_REGION --name $EKS_CLUSTER_NAME"
elif [ "$CLOUD" = "g" ]; then
  gcloud config set project "$GCP_PROJECT_ID"
  gcloud config set compute/zone "$GCP_COMPUTE_ZONE"
  CLUSTER_SETUP_COMMAND="gcloud container clusters get-credentials $GKE_CLUSTER_NAME"
fi

# 8. Generate GitHub Actions workflow
echo $'\n🔧 Step 8: Writing GitHub Actions deploy file'
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
    - name: Setup Kubernetes config
      run: $CLUSTER_SETUP_COMMAND
    - name: Deploy to Kubernetes
      run: kubectl apply -f k8s/
YAML

# Complete message
echo $'\n✅ Agentic AI Codespace setup complete.'
echo "- Auto-GPT is ready with local LLM via Ollama (mistral model)"
echo "- Cloud CLI configured and kubeconfig setup"
echo "- GitHub Actions ready to deploy to $([ "$CLOUD" = "a" ] && echo 'AWS EKS' || echo 'Google GKE')"
echo "\n👉 Push code to main branch to trigger Kubernetes deployment."
