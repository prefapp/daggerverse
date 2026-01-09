PORT=0
CREDENTIALS_FILE="./boot/CredentialsFile.yaml"
BOOTSTRAP_FILE="./boot/BootstrapFile.yaml"
# VOLUME_ID="<your volume cache id>"
CREATE_CLUSTER=false
CLUSTER_NAME=""
AUTO=false
LAST_EXIT_CODE=0
COMMAND_WAIT_TIME=5
DELETE_CLUSTER_ON_FAILURE=false

wait_for() {
    local WAIT_TIME=$1
    local SECS=0
    while [ $SECS -lt $WAIT_TIME ]; do
        echo "$((WAIT_TIME - SECS))..."
        sleep 1
        SECS=$((SECS + 1))
    done
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
        # Using built-in shell conversion is faster and avoids a subshell/external program
        RESPONSE=${RESPONSE,,}

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
    --auto | -a)
      AUTO=true
      shift # Move to the next argument
      ;;
    --kind-cluster-name | -k)
      CLUSTER_NAME="$2"
      shift 2 # Move past the flag AND its value
      ;;
    --help | -h)
      echo "Usage: $0 [--auto|-a] [--kind-cluster-name|-k <name>] [--wait-time|-w <seconds>] [--delete-cluster-on-failure|-d]"
      exit 0
      ;;
    *)
      # This captures unknown flags or positional arguments
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

if [ "$CLUSTER_NAME" == "" ]; then
    RANDOM_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c 8)
    CLUSTER_NAME="firestartr-kind-cluster-$RANDOM_SUFFIX"
    CREATE_CLUSTER=true
else
    PORT=$(docker inspect --format='{{(index (index .NetworkSettings.Ports "6443/tcp") 0).HostPort}}' $CLUSTER_NAME-control-plane)
fi



# Create kind cluster if needed
if [ "$CREATE_CLUSTER" == true ]; then
    if [ "$AUTO" == true ]; then
        echo "🤖 Auto mode enabled. Creating kind cluster ${CLUSTER_NAME} in..."
        wait_for $COMMAND_WAIT_TIME
        echo "Creating kind cluster ${CLUSTER_NAME}"
        ACTION="continue"
    else
        ACTION=$(prompt_continue_skip_abort "⚠️ Create new kind cluster ${CLUSTER_NAME}?")
    fi

    case "$ACTION" in
        "continue")
            kind create cluster --name "${CLUSTER_NAME}"
            LAST_EXIT_CODE=$?

            if [ "$LAST_EXIT_CODE" == 0 ]; then
                PORT=$(docker inspect --format='{{(index (index .NetworkSettings.Ports "6443/tcp") 0).HostPort}}' $CLUSTER_NAME-control-plane)
                echo "✅ Kind cluster ${CLUSTER_NAME} created. Port: ${PORT}."
            fi
            ;;
        "skip")
            CLUSTER_NAME="kind"
            echo "⏭️ Skipping the next section and moving to the end."
            ;;
        "abort")
            echo "🛑 Aborting script execution now."
            exit 1
            ;;
    esac
fi



# Validate bootstrap
if [ "$AUTO" == true ]; then
    if [ "$LAST_EXIT_CODE" != 0 ]; then
        echo "❌ Previous command failed with exit code $LAST_EXIT_CODE. Aborting."

        if [ "$DELETE_CLUSTER_ON_FAILURE" == true ]; then
            kind delete cluster --name "${CLUSTER_NAME}"
        fi

        exit $LAST_EXIT_CODE
    fi

    echo "🤖 Auto mode enabled. Validating bootstrap in..."
    wait_for $COMMAND_WAIT_TIME
    echo "Validating bootstrap"

    ACTION="continue"
else
    ACTION=$(prompt_continue_skip_abort "⚠️ Validate bootstrap?")
fi

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
        call cmd-validate-bootstrap \
        --kubeconfig="${HOME}/.kube" \
        --kind-svc=tcp://localhost:${PORT} \
        --kind-cluster-name="${CLUSTER_NAME}"

        LAST_EXIT_CODE=$?
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac


# Init secrets machinery
if [ "$AUTO" == true ]; then
    if [ "$LAST_EXIT_CODE" != 0 ]; then
        echo "❌ Previous command failed with exit code $LAST_EXIT_CODE. Aborting."

        if [ "$DELETE_CLUSTER_ON_FAILURE" == true ]; then
            kind delete cluster --name "${CLUSTER_NAME}"
        fi

        exit $LAST_EXIT_CODE
    fi

    echo "🤖 Auto mode enabled. Initializing secrets machinery in..."
    wait_for $COMMAND_WAIT_TIME
    echo "Initializing secrets machinery"

    ACTION="continue"
else
    ACTION=$(prompt_continue_skip_abort "⚠️ Init secret machinery?")
fi

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-init-secrets-machinery \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
       --kind-cluster-name="${CLUSTER_NAME}"

        LAST_EXIT_CODE=$?
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac


# Import and create basic CRs and Claims
if [ "$AUTO" == true ]; then
    if [ "$LAST_EXIT_CODE" != 0 ]; then
        echo "❌ Previous command failed with exit code $LAST_EXIT_CODE. Aborting."

        if [ "$DELETE_CLUSTER_ON_FAILURE" == true ]; then
            kind delete cluster --name "${CLUSTER_NAME}"
        fi

        exit $LAST_EXIT_CODE
    fi

    echo "🤖 Auto mode enabled. Importing existing resources and creating basic claims and CRs in..."
    wait_for $COMMAND_WAIT_TIME
    echo "Importing existing resources and creating basic claims and CRs"

    ACTION="continue"
else
    ACTION=$(prompt_continue_skip_abort "⚠️ Import and create the basic CRs and Claims?")
fi

case "$ACTION" in
    "continue")
dagger -vvv --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-import-resources \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
       --kind-cluster-name="${CLUSTER_NAME}" \
       --cache-volume=${VOLUME_ID}

        LAST_EXIT_CODE=$?
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac


# Push resources to the system's repos
if [ "$AUTO" == true ]; then
    if [ "$LAST_EXIT_CODE" != 0 ]; then
        echo "❌ Previous command failed with exit code $LAST_EXIT_CODE. Aborting."

        if [ "$DELETE_CLUSTER_ON_FAILURE" == true ]; then
            kind delete cluster --name "${CLUSTER_NAME}"
        fi

        exit $LAST_EXIT_CODE
    fi

    echo "🤖 Auto mode enabled. Pushing resources to state repositories in..."
    wait_for $COMMAND_WAIT_TIME
    echo "Pushing resources to state repositories"

    ACTION="continue"
else
    ACTION=$(prompt_continue_skip_abort "Push resources to the system's repos?")
fi

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-resources \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
       --kind-cluster-name="${CLUSTER_NAME}" \
       --cache-volume=${VOLUME_ID}

        LAST_EXIT_CODE=$?
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac


# Push state secrets
if [ "$AUTO" == true ]; then
    if [ "$LAST_EXIT_CODE" != 0 ]; then
        echo "❌ Previous command failed with exit code $LAST_EXIT_CODE. Aborting."

        if [ "$DELETE_CLUSTER_ON_FAILURE" == true ]; then
            kind delete cluster --name "${CLUSTER_NAME}"
        fi

        exit $LAST_EXIT_CODE
    fi

    echo "🤖 Auto mode enabled. Pushing organization state secrets in..."
    wait_for $COMMAND_WAIT_TIME
    echo "Pushing organization state secrets"

    ACTION="continue"
else
    ACTION=$(prompt_continue_skip_abort "Push organization state secrets (only for enterprise orgs)?")
fi

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-state-secrets \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
       --kind-cluster-name="${CLUSTER_NAME}" \
       --cache-volume=${VOLUME_ID}

        LAST_EXIT_CODE=$?
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac


# Push argocd - deployment
if [ "$AUTO" == true ]; then
    if [ "$LAST_EXIT_CODE" != 0 ]; then
        echo "❌ Previous command failed with exit code $LAST_EXIT_CODE. Aborting."

        if [ "$DELETE_CLUSTER_ON_FAILURE" == true ]; then
            kind delete cluster --name "${CLUSTER_NAME}"
        fi

        exit $LAST_EXIT_CODE
    fi

    echo "🤖 Auto mode enabled. Pushing argocd - deployment to the system's repos in..."
    wait_for $COMMAND_WAIT_TIME
    echo "Pushing argocd - deployment"

    ACTION="continue"
else
    ACTION=$(prompt_continue_skip_abort "Push argocd - deployment to the system's repos?")
fi

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-deployment

        LAST_EXIT_CODE=$?
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac


# Push argocd - permissions and secrets
if [ "$AUTO" == true ]; then
    if [ "$LAST_EXIT_CODE" != 0 ]; then
        echo "❌ Previous command failed with exit code $LAST_EXIT_CODE. Aborting."

        if [ "$DELETE_CLUSTER_ON_FAILURE" == true ]; then
            kind delete cluster --name "${CLUSTER_NAME}"
        fi

        exit $LAST_EXIT_CODE
    fi

    echo "🤖 Auto mode enabled. Pushing argocd - permissions and secrets to the system's repos..."
    wait_for $COMMAND_WAIT_TIME
    echo "Pushing argocd - permissions and secrets"

    ACTION="continue"
else
    ACTION=$(prompt_continue_skip_abort "Push argocd - permissions and secrets to the system's repos?")
fi

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-argo

        LAST_EXIT_CODE=$?
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac

if [ "$AUTO" == true ]; then
    if [ "$LAST_EXIT_CODE" != 0 ]; then
        echo "❌ Previous command failed with exit code $LAST_EXIT_CODE. Aborting."

        if [ "$DELETE_CLUSTER_ON_FAILURE" == true ]; then
            kind delete cluster --name "${CLUSTER_NAME}"
        fi

        exit $LAST_EXIT_CODE
    fi

    echo "🤖 Auto mode enabled. Deleting kind cluster ${CLUSTER_NAME} after the Bootstrap process has finished..."
    wait_for $COMMAND_WAIT_TIME
    echo "Deleting kind cluster ${CLUSTER_NAME}"

    ACTION="continue"
else
    ACTION=$(prompt_continue_skip_abort "Delete kind cluster ${CLUSTER_NAME}?")
fi

case "$ACTION" in
    "continue")
        kind delete cluster --name "${CLUSTER_NAME}"
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac

