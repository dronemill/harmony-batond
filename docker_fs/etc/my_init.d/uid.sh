#!/bin/sh

# http://chapeau.freevariable.com/2014/08/docker-uid.html

# if we don't have a UID to swich to, then get out
if [ -z "$EXEC_UID" ]; then exit; fi;

export ORIGPASSWD=$(cat /etc/passwd | grep $EXEC_UNAME)
export ORIG_UID=$(echo $ORIGPASSWD | cut -f3 -d:)
export ORIG_GID=$(echo $ORIGPASSWD | cut -f4 -d:)

export EXEC_UID=${EXEC_UID:=$ORIG_UID}
export EXEC_GID=${EXEC_GID:=$ORIG_GID}

ORIG_HOME=$(echo $ORIGPASSWD | cut -f6 -d:)

sed -i -e "s/:${ORIG_UID}:${ORIG_GID}:/:${EXEC_UID}:${EXEC_GID}:/" /etc/passwd
sed -i -e "s/${EXEC_UNAME}:x:${ORIG_GID}:/${EXEC_UNAME}:x:${EXEC_GID}:/" /etc/group

chown -R ${EXEC_UID}:${EXEC_GID} ${ORIG_HOME}
