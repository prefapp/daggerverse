#!/usr/bin/env bash
PORT=-1  # Populated later by the script
CREDENTIALS_FILE="./boot/CredentialsFile.yaml"
BOOTSTRAP_FILE="./boot/BootstrapFile.yaml"
VOLUME_ID="${VOLUME_ID:-}"
CLUSTER_NAME="${CLUSTER_NAME:-}"
CREATE_CLUSTER=false
AUTO=false
LAST_EXIT_CODE=0
COMMAND_WAIT_TIME=5
DELETE_CLUSTER_ON_FAILURE=false

wait_for() {
    local WAIT_TIME=$1
    for ((i=WAIT_TIME; i>0; i--)); do
        printf "\r⏱️  Starting in %d seconds... \e[K" "$i"
        sleep 1
    done
    printf "\r🚀 Starting now!\e[K\n"
}

wait_for_user() {
    echo ""
    echo "🛑 User intervention required. Please follow the instructions above."
    read -r -p "Press Enter to continue..."
    echo ""
}

prompt_continue_skip_abort() {
    local PROMPT_MSG="$1"
    local RESPONSE

    # Loop until a valid response is given
    while true; do
        # -p: Display the prompt message
        # -r: Prevents backslashes from being interpreted (safer)
        # -i: Provides a default value (not used here, but useful)
        read -r -p "$PROMPT_MSG [y(es)/n(o)/a(bort)]: " RESPONSE

        # --- NEW DEFAULT HANDLING ---
        # 1. If RESPONSE is empty (user just pressed Enter), set it to "y".
        RESPONSE=${RESPONSE:-y}
        # ----------------------------

        # Convert input to lowercase for case-insensitive comparison
        # Using tr is compatible with older bash versions (macOS ships with bash 3.2)
        RESPONSE=$(echo "$RESPONSE" | tr '[:upper:]' '[:lower:]')

        case "$RESPONSE" in
            "y" | "ye" | "yes")
                echo "continue"
                return 0
                ;;
            "n" | "no")
                echo "skip"
                return 0
                ;;
            "a" | "ab" | "abo" | "abor" | "abort")
                echo "abort"
                return 0
                ;;
            *)
                echo "❌ Invalid input. Valid values: 'y(es)', 'n(o)', or 'a(bort)'." >&2
                ;;
        esac
    done
}

check_dagger_version() {
    # Check if dagger is installed
    if ! command -v dagger &> /dev/null; then
        echo "❌ Dagger is not installed. Please install Dagger 0.19.7 or greater."
        echo "   Installation instructions: https://docs.dagger.io/install"
        exit 1
    fi

    # Get installed version
    local INSTALLED_VERSION
    INSTALLED_VERSION=$(dagger version 2>&1 | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1 | sed 's/v//')
    local MINIMUM_VERSION="0.19.7"

    # Compare versions (convert to comparable integers)
    local INSTALLED_MAJOR
    INSTALLED_MAJOR=$(echo "$INSTALLED_VERSION" | cut -d. -f1)
    local INSTALLED_MINOR
    INSTALLED_MINOR=$(echo "$INSTALLED_VERSION" | cut -d. -f2)
    local INSTALLED_PATCH
    INSTALLED_PATCH=$(echo "$INSTALLED_VERSION" | cut -d. -f3)
    
    local MINIMUM_MAJOR
    MINIMUM_MAJOR=$(echo "$MINIMUM_VERSION" | cut -d. -f1)
    local MINIMUM_MINOR
    MINIMUM_MINOR=$(echo "$MINIMUM_VERSION" | cut -d. -f2)
    local MINIMUM_PATCH
    MINIMUM_PATCH=$(echo "$MINIMUM_VERSION" | cut -d. -f3)

    # Version comparison logic (lexicographic: major, then minor, then patch)
    local VERSION_OK=false

    if [ "${INSTALLED_MAJOR}" -gt "${MINIMUM_MAJOR}" ]; then
        VERSION_OK=true
    elif [ "${INSTALLED_MAJOR}" -eq "${MINIMUM_MAJOR}" ]; then
        if [ "${INSTALLED_MINOR}" -gt "${MINIMUM_MINOR}" ]; then
            VERSION_OK=true
        elif [ "${INSTALLED_MINOR}" -eq "${MINIMUM_MINOR}" ]; then
            if [ "${INSTALLED_PATCH}" -ge "${MINIMUM_PATCH}" ]; then
                VERSION_OK=true
            fi
        fi
    fi

if [ "$VERSION_OK" = false ]; then
        echo "❌ Dagger version $INSTALLED_VERSION is installed, but version $MINIMUM_VERSION or greater is required."
        echo "   Please upgrade Dagger: https://docs.dagger.io/install"
        exit 1
    fi

    echo "✅ Dagger version $INSTALLED_VERSION detected (meets minimum requirement of $MINIMUM_VERSION)"
}

handle_command_failure() {
    local EXIT_CODE=$1
    
    if [ "$EXIT_CODE" -ne 0 ]; then
        echo "❌ Command failed with exit code $EXIT_CODE."
        
        if [ "$DELETE_CLUSTER_ON_FAILURE" = true ]; then
            echo "🗑️ Deleting kind cluster ${CLUSTER_NAME}..."
            kind delete cluster --name "${CLUSTER_NAME}"
        fi
        
        echo "🛑 Aborting script execution."
        exit "$EXIT_CODE"
    fi
}

prompt_or_auto() {
    local PROMPT_MSG="$1"
    local ACTION_DESC="$2"

    if [ "$AUTO" = true ]; then
        {
            echo "🤖 Auto: ${ACTION_DESC}"
            wait_for "$COMMAND_WAIT_TIME"
        } >&2

        echo "continue"
    else
        prompt_continue_skip_abort "$PROMPT_MSG"
    fi
}

execute_step() {
    local ACTION="$1"
    shift
    
    case "$ACTION" in
        "continue")
            "$@"
            LAST_EXIT_CODE=$?
            handle_command_failure "$LAST_EXIT_CODE"
            ;;
        "skip")
            echo "⏭️ Skipping the next section and moving to the end."
            ;;
        "abort")
            echo "🛑 Aborting script execution now."
            exit 1
            ;;
    esac
}

# Check Dagger version before proceeding
check_dagger_version

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --wait-time | -w)
            COMMAND_WAIT_TIME="$2"
            shift 2 # Move past the flag AND its value
            ;;
        --delete-cluster-on-failure | -d)
            DELETE_CLUSTER_ON_FAILURE=true
            shift # Move to the next argument
            ;;
        --auto-execute-script)
            AUTO=true
            shift # Move to the next argument
            ;;
        --kind-cluster-name | -k)
            CLUSTER_NAME="$2"
            shift 2 # Move past the flag AND its value
            ;;
        --help | -h)
            echo "Usage: $0 [--kind-cluster-name|-k <name>] [--delete-cluster-on-failure|-d] [--auto-execute-script] [--wait-time|-w <seconds>]"
            exit 0
            ;;
        *)
            # This captures unknown flags or positional arguments
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

if [ -z "$CLUSTER_NAME" ]; then
    RANDOM_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c 8)
    CLUSTER_NAME="firestartr-kind-cluster-$RANDOM_SUFFIX"
    CREATE_CLUSTER=true
else
    if ! PORT=$(docker inspect --format='{{(index (index .NetworkSettings.Ports "6443/tcp") 0).HostPort}}' "$CLUSTER_NAME"-control-plane); then
        echo "❌ Could not find existing kind cluster named ${CLUSTER_NAME}. Please check the name and try again."
        exit 1
    fi
fi



# Create kind cluster if needed
if [ "$CREATE_CLUSTER" = true ]; then
    if [ "$AUTO" = true ]; then
        echo "🤖 Auto: Creating kind cluster ${CLUSTER_NAME}"
        wait_for "$COMMAND_WAIT_TIME"
        ACTION="continue"
    else
        ACTION=$(prompt_continue_skip_abort "⚠️ Create new kind cluster ${CLUSTER_NAME}?")
    fi

    case "$ACTION" in
        "continue")
            kind create cluster --name "${CLUSTER_NAME}"
            LAST_EXIT_CODE=$?

            if [ "$LAST_EXIT_CODE" -eq 0 ]; then
                if ! PORT=$(docker inspect --format='{{(index (index .NetworkSettings.Ports "6443/tcp") 0).HostPort}}' "$CLUSTER_NAME"-control-plane); then
                    echo "❌ An error happened getting the port for cluster ${CLUSTER_NAME}. Please relaunch the script (you can use the flag '--kind-cluster-name ${CLUSTER_NAME}' to avoid creating a new cluster)"
                    exit 1
                fi
                echo "✅ Kind cluster ${CLUSTER_NAME} created. Port: ${PORT}."
            fi
            ;;
        "skip")
            echo "🛑 Skipping the cluster creation is not allowed. Please provide an existing cluster name via the --kind-cluster-name flag to skip this step"
            exit 1
            ;;
        "abort")
            echo "🛑 Aborting script execution now."
            exit 1
            ;;
    esac
fi



# Validate bootstrap
ACTION=$(prompt_or_auto "⚠️ Validate bootstrap?" "Validating bootstrap")
execute_step "$ACTION" dagger \
    --bootstrap-file="${BOOTSTRAP_FILE}" \
    --credentials-secret="file:${CREDENTIALS_FILE}" \
    call cmd-validate-bootstrap \
    --kubeconfig="${HOME}/.kube" \
    --kind-svc="tcp://localhost:${PORT}" \
    --kind-cluster-name="${CLUSTER_NAME}"


# Init secrets machinery
ACTION=$(prompt_or_auto "⚠️ Init secret machinery?" "Initializing secrets machinery")
execute_step "$ACTION" dagger \
    --bootstrap-file="${BOOTSTRAP_FILE}" \
    --credentials-secret="file:${CREDENTIALS_FILE}" \
    call cmd-init-secrets-machinery \
    --kubeconfig="${HOME}/.kube" \
    --kind-svc="tcp://localhost:${PORT}" \
    --kind-cluster-name="${CLUSTER_NAME}"


# Import and create basic CRs and Claims
ACTION=$(prompt_or_auto "⚠️ Import and create the basic CRs and Claims?" "Importing existing resources and creating basic claims and CRs")
execute_step "$ACTION" dagger \
    --bootstrap-file="${BOOTSTRAP_FILE}" \
    --credentials-secret="file:${CREDENTIALS_FILE}" \
    call cmd-import-resources \
    --kubeconfig="${HOME}/.kube" \
    --kind-svc="tcp://localhost:${PORT}" \
    --kind-cluster-name="${CLUSTER_NAME}" \
    --cache-volume="${VOLUME_ID}"


# Push resources to the system's repos
ACTION=$(prompt_or_auto "Push resources to the system's repos?" "Pushing resources to state repositories")
execute_step "$ACTION" dagger \
    --bootstrap-file="${BOOTSTRAP_FILE}" \
    --credentials-secret="file:${CREDENTIALS_FILE}" \
    call cmd-push-resources \
    --kubeconfig="${HOME}/.kube" \
    --kind-svc="tcp://localhost:${PORT}" \
    --kind-cluster-name="${CLUSTER_NAME}" \
    --cache-volume="${VOLUME_ID}"


# Push state secrets
ACTION=$(prompt_or_auto "Push organization state secrets (only for non-free orgs)?" "Pushing organization state secrets")
execute_step "$ACTION" dagger \
    --bootstrap-file="${BOOTSTRAP_FILE}" \
    --credentials-secret="file:${CREDENTIALS_FILE}" \
    call cmd-push-state-secrets \
    --kubeconfig="${HOME}/.kube" \
    --kind-svc="tcp://localhost:${PORT}" \
    --kind-cluster-name="${CLUSTER_NAME}" \
    --cache-volume="${VOLUME_ID}"


# Push argocd - deployment
ACTION=$(prompt_or_auto "Push argocd - deployment to the system's repos?" "Pushing argocd - deployment")
execute_step "$ACTION" dagger \
    --bootstrap-file="${BOOTSTRAP_FILE}" \
    --credentials-secret="file:${CREDENTIALS_FILE}" \
    call cmd-push-deployment

if [ "$ACTION" = "continue" ]; then
    wait_for_user
fi


# Push argocd - permissions and secrets
ACTION=$(prompt_or_auto "Push argocd - permissions and secrets to the system's repos?" "Pushing argocd - permissions and secrets")
execute_step "$ACTION" dagger \
    --bootstrap-file="${BOOTSTRAP_FILE}" \
    --credentials-secret="file:${CREDENTIALS_FILE}" \
    call cmd-push-argo

if [ "$ACTION" = "continue" ]; then
    wait_for_user
fi

# Delete cluster after successful completion
ACTION=$(prompt_or_auto "Delete kind cluster ${CLUSTER_NAME}?" "Deleting kind cluster ${CLUSTER_NAME} after the Bootstrap process has finished")
execute_step "$ACTION" kind delete cluster --name "${CLUSTER_NAME}"

echo "✨ Bootstrap process completed successfully! ✨"
