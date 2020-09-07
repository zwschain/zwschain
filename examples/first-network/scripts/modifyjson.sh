#!/bin/bash

function ModifyJson()
{

filename=../node1/ptn-genesis.json
pk=$5
index=$[ $4 - 1 ]

add=`cat $filename | jq ".initialMediatorCandidates[$index] |= . + {\"account\": \"$1\", \"rewardAdd\": \"$1\", \"initPubKey\": \"$2\", \"node\": \"$3\",\"public_key\": \"$pk\", \"reward_address\": \"$1\"}"`

if [ $index -eq 0 ] ; then

    createaccount=`./createaccount.sh`
    account=`echo $createaccount | sed -n '$p'| awk '{print $NF}'`
    account=${account:0:35}
    account=`echo ${account///}`

    add=`echo $add | jq ".tokenHolder = \"$account\""`
    add=`echo $add | jq ".digitalIdentityConfig.rootCAHolder = \"$account\""`

    createaccount=`./createaccount.sh`
    account=`echo $createaccount | sed -n '$p'| awk '{print $NF}'`
    account=${account:0:35}
    account=`echo ${account///}`

    add=`echo $add | jq ".initialParameters.foundation_address = \"$account\""`

fi

    rm $filename
    echo $add >> temp.json
    jq -r . temp.json >> $filename
    rm temp.json
}

