#!/bin/bash

#set -x

if [ "$#" -ne 3 ]
then
    echo -e "$0 <username> <password> <base_url:port>"
    exit 1
fi

USERNAME="$1"
PASSWORD="$2"

BASE_URL="$3"
LOGIN_URL="${BASE_URL}/login"
LOGOUT_URL="${BASE_URL}/logout"
CONFIGS_URL="${BASE_URL}/configs/"

COOKIE_FILE="cookies.txt"
SAVE_COOKIE="-c ${COOKIE_FILE}"
USE_COOKIE="-b ${COOKIE_FILE}"
HEADER="Content-Type: application/json"

read -r -d '' LOGINDATA <<EOF
{
    "id": 1,
    "username": "$1",
    "password": "$2"
}
EOF

read -r -d '' DATA <<'EOF'
{
    "id": 2,
    "configurations": [
        {
            "name": "workstation1",
            "hostname": "e305.lab.com",
            "port": 1025,
            "username": "rain",
            "location": "hq"
        },
        {
            "name": "workstation2",
            "hostname": "x205.lab.com",
            "port": 1025,
            "username": "africa",
            "location": "remoteA"
        },
        {
            "name": "thinclient9",
            "hostname": "e304t.lab.com",
            "port": 45873,
            "username": "rain",
            "location": "remoteA"
        },
        {
            "name": "thinclient2",
            "hostname": "x204t.lab.com",
            "port": 35263,
            "username": "africa",
            "location": "remoteA"
        },
        {
            "name": "host1",
            "hostname": "nexus-ntp.lab.com",
            "port": 1241,
            "username": "toto",
            "location": "remoteA"
        },
        {
            "name": "host2",
            "hostname": "nexus-xml.lab.com",
            "port": 3384,
            "username": "admin"
        }
    ]
}
EOF

read -r -d '' DELETEDATA <<'EOF'
{
    "id": 3,
    "configurations": [
        {
            "name": "workstation2",
            "hostname": "x205.lab.com",
            "port": 35263,
            "username": "africa"
        },
        {
            "name": "host1",
            "hostname": "nexus-ntp.lab.com",
            "port": 1241,
            "username": "toto"
        }
    ]
}
EOF

function lots_of_kv {
    for num in $(seq $1)
    do
        echo -e "\"key$num\": $num,"
    done
}

read -r -d '' BIGDATA <<EOF
{
    "id": 4,
    "configurations": [
        {
            $(lots_of_kv 101)
            "name": "cray500",
            "hostname": "x205.lab.com",
            "port": 35263,
            "username": "africa"
        },
        {
            $(lots_of_kv 111)
            "name": "cray50",
            "hostname": "nexus-ntp.lab.com",
            "port": 1241,
            "username": "toto"
        }
    ]
}
EOF

echo -e "\n"
echo "Removing ${COOKIE_FILE}"
rm ${COOKIE_FILE}

echo -e "\n"
echo "Logging in"
echo "**************************************************"
curl -i ${SAVE_COOKIE} -H "${HEADER}" -X POST -d "${LOGINDATA}" ${LOGIN_URL}

echo -e "\n"
echo "Getting configs"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X GET ${CONFIGS_URL}

echo -e "\n"
echo "Adding configs"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X POST -d "${DATA}" ${CONFIGS_URL}

echo -e "\n"
echo "Getting configs"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X GET ${CONFIGS_URL}

echo -e "\n"
echo "Getting configs sorted by username, then hostname"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X GET ${CONFIGS_URL}username/hostname

echo -e "\n"
echo "Getting configs sorted by port"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X GET ${CONFIGS_URL}port

echo -e "\n"
echo "Deleting configs"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X DELETE -d "${DELETEDATA}" ${CONFIGS_URL}

echo -e "\n"
echo "Getting configs"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X GET ${CONFIGS_URL}

echo -e "\n"
echo "Getting configs with more than 100 entries"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X GET ${CONFIGS_URL}large

echo -e "\n"
echo "Adding LARGE configs"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X POST -d "${BIGDATA}" ${CONFIGS_URL}

echo -e "\n"
echo "Getting configs with more than 100 entries"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X GET ${CONFIGS_URL}large

echo -e "\n"
echo "Logging out"
echo "**************************************************"
curl -i ${SAVE_COOKIE} -H "${HEADER}" -X POST ${LOGOUT_URL}

echo -e "\n"
echo "Getting configs"
echo "**************************************************"
curl -i ${USE_COOKIE} -H "${HEADER}" -X GET ${CONFIGS_URL}

