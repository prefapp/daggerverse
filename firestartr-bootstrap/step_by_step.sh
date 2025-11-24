PORT=<kind cluster port>
CREDENTIALS_FILE="./boot/CredentialsFile.yaml"
VOLUME_ID="<your volume cache id>"

function prompt_continue_skip_abort() {
    local PROMPT_MSG="$1"
    local RESPONSE

    # Loop until a valid response is given
    while true; do
        # -p: Display the prompt message
        # -r: Prevents backslashes from being interpreted (safer)
        # -i: Provides a default value (not used here, but useful)
        read -r -p "$PROMPT_MSG [continue/skip/abort]: " RESPONSE

        # Convert input to lowercase for case-insensitive comparison
        RESPONSE=$(echo "$RESPONSE" | tr '[:upper:]' '[:lower:]')

        case "$RESPONSE" in
            "continue")
                echo "continue"
                return 0
                ;;
            "skip")
                echo "skip"
                return 0
                ;;
            "abort")
                echo "abort"
                return 0
                ;;
            *)
                echo "❌ Invalid input. Please enter 'continue', 'skip', or 'abort'." >&2
                ;;
        esac
    done
}

ACTION=$(prompt_continue_skip_abort "⚠️ Validate bootstrap?")

case "$ACTION" in
    "continue")
dagger --bootstrap-file="./boot/BoostrapFile.yaml" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-validate-bootstrap 
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
dagger --bootstrap-file="./boot/BoostrapFile.yaml" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-init-secrets-machinery \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} 
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
dagger --bootstrap-file="./boot/BoostrapFile.yaml" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-import-resources \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
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
dagger --bootstrap-file="./boot/BoostrapFile.yaml" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-resources \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
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
dagger --bootstrap-file="./boot/BoostrapFile.yaml" \
       --credentials-secret="file:${CREDENTIALS_FILE}" \
       call cmd-push-state-secrets \
       --kubeconfig="${HOME}/.kube" \
       --kind-svc=tcp://localhost:${PORT} \
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
dagger --bootstrap-file="./boot/BoostrapFile.yaml" \
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
dagger --bootstrap-file="./boot/BoostrapFile.yaml" \
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
