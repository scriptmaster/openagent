#!/bin/bash

# Check for profile flag
PROFILE="default"
if [[ "$1" == "--profile" && -n "$2" ]]; then
  PROFILE="$2"
fi

echo "🛠 Running setup in profile: $PROFILE"

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
  echo $'\n🔧 Installing kubectl (curl-based method)'
  curl -LO "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  chmod +x kubectl
  sudo mv kubectl /usr/local/bin/
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

echo $'\n🔧 Step 5: Setting up Auto-GPT (Docker-based)'

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

# Use profile for server optimizations
COMPOSE_PROFILE_ARG=""
if [ "$PROFILE" = "server" ]; then
  COMPOSE_PROFILE_ARG="--profile server"
fi

# Check if any containers are already running
if docker compose ps | grep -q 'Up'; then
  echo "✅ Docker containers are running."
  confirm "Rebuild and restart containers (slower)?" && docker compose up -d --build $COMPOSE_PROFILE_ARG || echo "⚡ Running containers left as-is."
else
  docker compose up -d --build $COMPOSE_PROFILE_ARG
fi

if [ -d "frontend" ]; then
  cd frontend
  cp .env.example .env
  npm install
  npm run dev
  cd ..
else
  echo "⚠️  Frontend folder not found — skipping UI setup."
fi

cd ../..
if [ ! -f ".env" ]; then
  cat << EOF > .env
OPEN
::contentReference[oaicite:2]{index=2}
 

