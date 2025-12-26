#!/bin/bash

# OpenShift API Curl Cheatsheet
# This tool demonstrates curl commands for OpenShift API operations
# Based on the golang-client example

# Colors for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Auto-detect values from oc command or use environment variables
get_oc_token() {
    if command -v oc &> /dev/null; then
        oc whoami -t 2>/dev/null || echo ""
    else
        echo ""
    fi
}

get_oc_server() {
    if command -v oc &> /dev/null; then
        oc whoami --show-server 2>/dev/null || echo "https://api.cluster.example.com:6443"
    else
        echo "https://api.cluster.example.com:6443"
    fi
}

get_oc_username() {
    if command -v oc &> /dev/null; then
        oc whoami 2>/dev/null || echo "your-username"
    else
        echo "your-username"
    fi
}

# Use auto-detected values or environment variables as fallback
OPENSHIFT_SERVER=${OPENSHIFT_SERVER:-$(get_oc_server)}
USERNAME=${USERNAME:-$(get_oc_username)}
TOKEN=${TOKEN:-$(get_oc_token)}

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${CYAN}     OpenShift API Curl Cheatsheet     ${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${YELLOW}Server:${NC} $OPENSHIFT_SERVER"
    echo -e "${YELLOW}Username:${NC} $USERNAME"
    echo -e "${YELLOW}Token:${NC} ${TOKEN:0:20}..."
    echo ""
}

print_menu() {
    echo -e "${GREEN}Available Operations:${NC}"
    echo "1. Create User"
    echo "2. Create Group"
    echo "3. Add User to Group"
    echo "4. Deploy Model (LLMInferenceService)"
    echo "5. List LLMInferenceServices"
    echo "X. Exit"
    echo ""
    echo -e "${BLUE}========================================${NC}"
}

get_user_input() {
    local prompt="$1"
    local default="$2"
    local input
    
    echo -e "${YELLOW}$prompt${NC}"
    if [ ! -z "$default" ]; then
        echo -n "(default: $default): "
    else
        echo -n ": "
    fi
    read -r input
    if [ -z "$input" ] && [ ! -z "$default" ]; then
        input="$default"
    fi
    echo "$input"
}

print_curl_command() {
    local title="$1"
    local description="$2"
    local curl_command="$3"
    
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}$title${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${YELLOW}Description:${NC}"
    echo "$description"
    echo ""
    echo -e "${YELLOW}Curl Command:${NC}"
    # Use printf with literal escape sequences that work reliably
    printf "${CYAN}%s${NC}\n" "$curl_command" | sed "s|<\([^>]*\)>|$(printf "\033[0m\033[1;33m")<\1>$(printf "\033[0m\033[0;36m")|g"
    echo ""
    echo -e "${YELLOW}To use this command:${NC}"
    echo -e "1. Replace the placeholders in ${YELLOW}<>${NC} with your actual values"
    echo "2. Copy and paste the command into your terminal"
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

create_user_demo() {
    local curl_cmd="curl -X POST \\
  -H \"Authorization: Bearer \$(oc whoami -t)\" \\
  -H \"Content-Type: application/json\" \\
  -H \"Accept: application/json\" \\
  -k \\
  \"\$OPENSHIFT_SERVER/apis/user.openshift.io/v1/users\" \\
  -d '{
    \"apiVersion\": \"user.openshift.io/v1\",
    \"kind\": \"User\",
    \"metadata\": {
      \"name\": \"<USER_NAME>\"
    }
  }'"

    print_curl_command \
        "CREATE USER" \
        "Creates a new user in OpenShift using the user.openshift.io/v1 API" \
        "$curl_cmd"
}

add_user_to_group_demo() {
    echo ""
    echo -e "${YELLOW}⚠️  IMPORTANT: No direct 'add user' API - OpenShift Groups use declarative model${NC}"
    echo -e "${YELLOW}The users field represents the complete desired state, not an append operation${NC}"
    echo ""
    
    local get_cmd="curl -X GET \\
  -H \"Authorization: Bearer \$(oc whoami -t)\" \\
  -H \"Accept: application/json\" \\
  -k \\
  \"\$OPENSHIFT_SERVER/apis/user.openshift.io/v1/groups/<GROUP_NAME>\""

    print_curl_command \
        "STEP 1: GET EXISTING GROUP" \
        "First, get the current group to see existing users" \
        "$get_cmd"
    
    local patch_cmd="curl -X PATCH \\
  -H \"Authorization: Bearer \$(oc whoami -t)\" \\
  -H \"Content-Type: application/merge-patch+json\" \\
  -H \"Accept: application/json\" \\
  -k \\
  \"\$OPENSHIFT_SERVER/apis/user.openshift.io/v1/groups/<GROUP_NAME>\" \\
  -d '{
    \"users\": [\"<EXISTING_USER1>\", \"<EXISTING_USER2>\", \"<NEW_USER>\"]
  }'"

    print_curl_command \
        "STEP 2: PATCH WITH ALL USERS" \
        "Update the group with ALL users (existing + new) to avoid deleting existing users" \
        "$patch_cmd"
    
    echo -e "${RED}⚠️  WARNING: Using PATCH with just [\"$username\"] will REPLACE all existing users!${NC}"
    echo -e "${YELLOW}Safe approach:${NC} Always include existing users + the new user in the users array"
    echo ""
    echo -e "${CYAN}Alternative: Use 'oc' command for safer group management:${NC}"
    echo -e "${CYAN}   oc adm groups add-users <_Group_Name> <_User_Name>${NC}"
    echo ""
    echo -e "${YELLOW}Note:${NC} This assumes the group already exists. To create a group first, use option 2."
    echo ""
}

deploy_model_demo() {
    local curl_cmd="curl -X POST \\
  -H \"Authorization: Bearer \$(oc whoami -t)\" \\
  -H \"Content-Type: application/json\" \\
  -H \"Accept: application/json\" \\
  -k \\
  \"\$OPENSHIFT_SERVER/apis/serving.kserve.io/v1alpha1/namespaces/<NAMESPACE>/llminferenceservices\" \\
  -d '{
    \"apiVersion\": \"serving.kserve.io/v1alpha1\",
    \"kind\": \"LLMInferenceService\",
    \"metadata\": {
      \"annotations\": {
        \"alpha.maas.opendatahub.io/tiers\": \"[\\\"redhat-users-tier\\\"]\"
      },
      \"name\": \"<MODEL_NAME>\",
      \"namespace\": \"<NAMESPACE>\"
    },
    \"spec\": {
      \"model\": {
        \"name\": \"facebook/opt-125m\",
        \"uri\": \"hf://facebook/opt-125m\"
      },
      \"replicas\": 1,
      \"router\": {
        \"gateway\": {
          \"refs\": [
            {
              \"name\": \"maas-default-gateway\",
              \"namespace\": \"openshift-ingress\"
            }
          ]
        },
        \"route\": {}
      },
      \"template\": {
        \"containers\": [
          {
            \"args\": [
              \"--port\", \"8000\",
              \"--model\", \"facebook/opt-125m\",
              \"--mode\", \"random\",
              \"--ssl-certfile\", \"/var/run/kserve/tls/tls.crt\",
              \"--ssl-keyfile\", \"/var/run/kserve/tls/tls.key\"
            ],
            \"command\": [\"/app/llm-d-inference-sim\"],
            \"env\": [
              {
                \"name\": \"POD_NAME\",
                \"valueFrom\": {
                  \"fieldRef\": {
                    \"apiVersion\": \"v1\",
                    \"fieldPath\": \"metadata.name\"
                  }
                }
              },
              {
                \"name\": \"POD_NAMESPACE\",
                \"valueFrom\": {
                  \"fieldRef\": {
                    \"apiVersion\": \"v1\",
                    \"fieldPath\": \"metadata.namespace\"
                  }
                }
              }
            ],
            \"image\": \"ghcr.io/llm-d/llm-d-inference-sim:v0.5.1\",
            \"imagePullPolicy\": \"Always\",
            \"livenessProbe\": {
              \"httpGet\": {
                \"path\": \"/health\",
                \"port\": \"https\",
                \"scheme\": \"HTTPS\"
              }
            },
            \"name\": \"main\",
            \"ports\": [
              {
                \"containerPort\": 8000,
                \"name\": \"https\",
                \"protocol\": \"TCP\"
              }
            ],
            \"readinessProbe\": {
              \"httpGet\": {
                \"path\": \"/ready\",
                \"port\": \"https\",
                \"scheme\": \"HTTPS\"
              }
            }
          }
        ]
      }
    }
  }'"

    print_curl_command \
        "DEPLOY MODEL (LLMInferenceService)" \
        "Deploys an LLMInferenceService model using the KServe API (serving.kserve.io/v1alpha1)" \
        "$curl_cmd"
    
    echo -e "${YELLOW}Notes:${NC}"
    echo "• This deploys the facebook/opt-125m model as configured in the Go client example"
    echo "• The namespace must exist before deploying"
    echo "• Model will be accessible via the maas-default-gateway route"
    echo "• You can modify the model name/uri, replicas, and other parameters as needed"
    echo ""
}

list_users_demo() {
    local curl_cmd="curl -X GET \\
  -H \"Authorization: Bearer \$(oc whoami -t)\" \\
  -H \"Accept: application/json\" \\
  -k \\
  \"\$OPENSHIFT_SERVER/apis/user.openshift.io/v1/users\""

    print_curl_command \
        "LIST USERS" \
        "Retrieves a list of all users in the OpenShift cluster" \
        "$curl_cmd"
}

get_user_demo() {
    local curl_cmd="curl -X GET \\
  -H \"Authorization: Bearer \$(oc whoami -t)\" \\
  -H \"Accept: application/json\" \\
  -k \\
  \"\$OPENSHIFT_SERVER/apis/user.openshift.io/v1/users/<USER_NAME>\""

    print_curl_command \
        "GET USER DETAILS" \
        "Retrieves detailed information about a specific user" \
        "$curl_cmd"
}

delete_user_demo() {
    local curl_cmd="curl -X DELETE \\
  -H \"Authorization: Bearer \$(oc whoami -t)\" \\
  -H \"Accept: application/json\" \\
  -k \\
  \"\$OPENSHIFT_SERVER/apis/user.openshift.io/v1/users/<USER_NAME>\""

    print_curl_command \
        "DELETE USER" \
        "Deletes a user from the OpenShift cluster (WARNING: This action cannot be undone!)" \
        "$curl_cmd"
    
    echo -e "${RED}⚠️  WARNING: This will permanently delete the user!${NC}"
    echo ""
}

list_groups_demo() {
    local curl_cmd="curl -X GET \\
  -H \"Authorization: Bearer \$(oc whoami -t)\" \\
  -H \"Accept: application/json\" \\
  -k \\
  \"\$OPENSHIFT_SERVER/apis/user.openshift.io/v1/groups\""

    print_curl_command \
        "LIST GROUPS" \
        "Retrieves a list of all groups in the OpenShift cluster" \
        "$curl_cmd"
}

create_group_demo() {
    local curl_cmd="curl -X POST \\
  -H \"Authorization: Bearer \$(oc whoami -t)\" \\
  -H \"Content-Type: application/json\" \\
  -H \"Accept: application/json\" \\
  -k \\
  \"\$OPENSHIFT_SERVER/apis/user.openshift.io/v1/groups\" \\
  -d '{
    \"apiVersion\": \"user.openshift.io/v1\",
    \"kind\": \"Group\",
    \"metadata\": {
      \"name\": \"<GROUP_NAME>\"
    },
    \"users\": []
  }'"

    print_curl_command \
        "CREATE GROUP" \
        "Creates a new group in OpenShift using the user.openshift.io/v1 API" \
        "$curl_cmd"
}

show_token_info() {
    echo ""
    echo -e "${YELLOW}📝 Getting Your Bearer Token:${NC}"
    echo ""
    echo "To get your bearer token, follow these steps:"
    echo ""
    echo "1. Login to the OpenShift web console:"
    echo -e "${CYAN}   https://console-openshift-console.apps.ocp.example.com${NC}"
    echo ""
    echo "2. Click on your username in the top-right corner"
    echo ""
    echo "3. Select 'Copy login command'"
    echo ""
    echo "4. Click 'Display Token' and copy the 'oc login' command"
    echo ""
    echo "5. Paste and run the 'oc login' command in your terminal"
    echo ""
    echo "6. Confirm you're logged in:"
    echo -e "${CYAN}   oc whoami${NC}"
    echo ""
    echo "7. Set the server environment variable:"
    echo -e "${CYAN}   export OPENSHIFT_SERVER=\"https://api.ocp.example.com:6443\"${NC}"
    echo ""
    echo "8. Run this script again - it will auto-detect your credentials"
    echo ""
    
    if [ -z "$TOKEN" ] || [ "$TOKEN" = "" ]; then
        echo -e "${YELLOW}⚠️  No authentication detected! Please follow the steps above.${NC}"
        echo ""
    else
        echo -e "${GREEN}✅ Authentication auto-detected from 'oc' command${NC}"
        echo ""
    fi
}

main() {
    while true; do
        clear
        print_header
        show_token_info
        print_menu
        
        echo -n -e "${YELLOW}Select an option: ${NC}"
        read -r choice
        
        case $choice in
            1)
                create_user_demo
                ;;
            2)
                create_group_demo
                ;;
            3)
                add_user_to_group_demo
                ;;
            4)
                deploy_model_demo
                ;;
            5)
                list_llminferenceservices_demo
                ;;
            [Xx])
                echo -e "${GREEN}Thanks for using the OpenShift API Curl Cheatsheet!${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}Invalid option. Please try again.${NC}"
                sleep 2
                continue
                ;;
        esac
        
        echo ""
        echo -e "${BLUE}Press Enter to continue...${NC}"
        read -r
    done
}

# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi