PORT=0
CREDENTIALS_FILE="./boot/CredentialsFile.yaml"
BOOTSTRAP_FILE="./boot/BootstrapFile.yaml"
# VOLUME_ID="<your volume cache id>"
RANDOM_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c 8)
CLUSTER_NAME="firestartr-kind-cluster-$RANDOM_SUFFIX"

function prompt_continue_skip_abort() {
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

ACTION=$(prompt_continue_skip_abort "⚠️ Create new kind cluster ${CLUSTER_NAME}?")

case "$ACTION" in
    "continue")
        kind create cluster --name "${CLUSTER_NAME}"
        PORT=$(docker inspect --format='{{(index (index .NetworkSettings.Ports "6443/tcp") 0).HostPort}}' $CLUSTER_NAME-control-plane)
        echo "✅ Kind cluster ${CLUSTER_NAME} created. Port: ${PORT}."
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

ACTION=$(prompt_continue_skip_abort "⚠️ Validate bootstrap?")

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
        call cmd-validate-bootstrap \
        --kubeconfig="${HOME}/.kube" \
        --kind-svc=tcp://localhost:${PORT} \
        --kind-cluster-name="${CLUSTER_NAME}"
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac

ACTION=$(prompt_continue_skip_abort "⚠️ Init secret machinery?")

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-init-secrets-machinery \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
       --kind-cluster-name="${CLUSTER_NAME}"
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac

ACTION=$(prompt_continue_skip_abort "⚠️ Import and create the basic CRs and Claims?")

case "$ACTION" in
    "continue")
dagger -vvv --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-import-resources \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
       --kind-cluster-name="${CLUSTER_NAME}" \
       --cache-volume=${VOLUME_ID}
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac

ACTION=$(prompt_continue_skip_abort "Push resources to the system's repos?")

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-resources \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
       --kind-cluster-name="${CLUSTER_NAME}" \
       --cache-volume=${VOLUME_ID}
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac

ACTION=$(prompt_continue_skip_abort "Push organization state secrets (only for enterprise orgs)?")

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-state-secrets \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
       --kind-cluster-name="${CLUSTER_NAME}" \
       --cache-volume=${VOLUME_ID}
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac

ACTION=$(prompt_continue_skip_abort "Push argocd - deployment to the system's repos?")

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-deployment
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac

ACTION=$(prompt_continue_skip_abort "Push argocd - permissions and secrets to the system's repos?")

case "$ACTION" in
    "continue")
dagger --bootstrap-file="${BOOTSTRAP_FILE}" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-argo
        ;;
    "skip")
        echo "⏭️ Skipping the next section and moving to the end."
        ;;
    "abort")
        echo "🛑 Aborting script execution now."
        exit 1
        ;;
esac

ACTION=$(prompt_continue_skip_abort "Delete kind cluster ${CREATE_CLUSTER}?")

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
