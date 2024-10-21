#!/bin/bash

set -eo pipefail

# Default values (can be overridden by environment variables)
IMAGE_NAME="${IMAGE_NAME:-auth-service}"
TAG="${TAG:-0.0.5}"
PROJECT_ID="${PROJECT_ID:-composed-circle-437303-d2}"
REPOSITORY="${REPOSITORY:-service}"
LOCATION="${LOCATION:-asia-east1}"
DOCKERFILE_DIR="${DOCKERFILE_DIR:-.}"
PLATFORM="${PLATFORM:-linux/amd64}"
LOG_FILE="${LOG_FILE:-build_push.log}"

# Function to log messages
log_message() {
  echo "$(date +'%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Function to display help message
show_help() {
  echo "Usage: $0 [OPTIONS]"
  echo "Options:"
  echo "  --tag <tag>        Specify the image tag (default: $TAG)"
  echo "  --project <id>     Specify the project ID (default: $PROJECT_ID)"
  echo "  --help             Display this help message"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --tag)
      TAG="$2"
      shift 2
      ;;
    --project)
      PROJECT_ID="$2"
      shift 2
      ;;
    --help)
      show_help
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      show_help
      exit 1
      ;;
  esac
done

# Create or clear the log file
# shellcheck disable=SC2188
> "$LOG_FILE"

# Trap for error handling
trap 'log_message "An error occurred. Exiting."; exit 1' ERR

# Function to build image
build_image() {
  log_message "Building the Podman image..."
  podman build --platform "$PLATFORM" -t "${IMAGE_NAME}:${TAG}" "${DOCKERFILE_DIR}"
  log_message "Image ${IMAGE_NAME}:${TAG} built successfully."
}

# Function to tag image
tag_image() {
  log_message "Tagging the image for Artifact Registry..."
  podman tag "${IMAGE_NAME}:${TAG}" "${LOCATION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY}/${IMAGE_NAME}:${TAG}"
  log_message "Image tagged successfully."
}

# Function to push image
push_image() {
  log_message "Pushing the image to Artifact Registry..."
  podman push "${LOCATION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY}/${IMAGE_NAME}:${TAG}"
  log_message "Image pushed to Google Artifact Registry successfully."
}

# Function to check if version exists
check_version() {
  log_message "Checking if version already exists..."
  if podman manifest inspect "${LOCATION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY}/${IMAGE_NAME}:${TAG}" &> /dev/null; then
    log_message "Version ${TAG} already exists in the registry."
    read -p "Do you want to overwrite? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      log_message "Aborting push operation."
      exit 1
    fi
  fi
}

# Function to clean up
cleanup() {
  log_message "Cleaning up..."
  podman rmi "${IMAGE_NAME}:${TAG}" "${LOCATION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY}/${IMAGE_NAME}:${TAG}"
  log_message "Cleanup completed."
}

# Main execution
build_image
tag_image
check_version
push_image

# Optional cleanup
read -p "Do you want to remove local images? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  cleanup
fi

log_message "Process completed successfully."