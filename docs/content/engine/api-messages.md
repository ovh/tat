---
weight: 2
toc: true
title: "API - Messages"
prev: "/engine/general"
next: "/engine/api-topics"

---

## Store a new message

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "text": "text" }' \
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

You can add labels from the creation

```bash
curl -XPOST \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "text": "text", "dateCreation": 11123232, "labels": [{"text": "labelA", "color": "#eeeeee"}, {"text": "labelB", "color": "#ffffff"}] }' \
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

You can add replies from the creation

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
  -d '{ "text": "text", "replies":["reply A", "reply B"] }' \
  https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

You can add replies, with labels, from the creation

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
  -d '{ "text": "text msg root", "messages": [{ "text": "text reply", "labels": [{"text": "labelA", "color": "#eeeeee"}] }] }' \
  https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

If you use a `system user`, you can force message's date

```bash
curl -XPOST \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "text": "text", "dateCreation": 11123232 }' \
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

Return HTTP 201 if OK

## Store some messages

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '[{ "text": "text message A" },{ "text": "text message B", "labels": [{"text": "labelA", "color": "#eeeeee"}] }]' \
	https://<tatHostname>:<tatPort>/messages/topic/sub-topic
```

## Action on a existing message

Reply, Like, Unlike, Add Label, Remove Label, etc... use idReference but it's possible to use :

* TagReference
* StartTagReference
* LabelReference
* StartLabelReference

```bash
curl -XPOST \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "text": "text", "startTagReference": "keyTag:", "action": "reply"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

If several messages matche to your request, Tat gives you a HTTP Bad Request.

## Reply to a message

```bash
curl -XPOST \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "text": "text", "idReference": "9797q87KJhqsfO7Usdqd", "action": "reply"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Like a message
```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "like"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Unlike a message
```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "unlike"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Add a label to a message
*option* is the background color of the label.

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "label", "text": "myLabel", "option": "rgba(143,199,148,0.61)"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Remove a label from a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "unlabel", "text": "myLabel"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Remove all labels and add new ones

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "relabel", "labels": [{"text": "labelA", "color": "#eeeeee"}, {"text": "labelB", "color": "#ffffff"}]}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

Return HTTP 201 if OK

## Remove some labels and add new ones

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "relabel", "labels": [{"text": "labelA", "color": "#eeeeee"}, {"text": "labelB", "color": "#ffffff"}], "options": ["labelAToRemove", "labelAToRemove"] }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Update a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "update", "text": "my New Mesage updated"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Concat a message : adding additional text to one message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "concat", "text": " additional text"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Move a message to another topic

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "move", "option": "/newTopic/subNewTopic"}'\
	https://<tatHostname>:<tatPort>/message/oldTOpic/oldSubTopic
```

## Delete a message
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/message/nocascade/9797q87KJhqsfO7Usdqd/topic/subTopic
```

## Delete a message and its replies
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/message/cascade/9797q87KJhqsfO7Usdqd/topic/subTopic
```

## Delete a message and its replies, even if it's in Tasks Topic of one user
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/message/cascadeforce/9797q87KJhqsfO7Usdqd/topic/subTopic
```

## Delete a list of messages
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/messages/nocascade/topic/subTopic?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name
```

see https://github.com/ovh/tat#parameters for all parameters

## Delete a list of messages and its replies
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/messages/cascade/topic/subTopic?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name
```

see https://github.com/ovh/tat#parameters for all parameters

## Delete a list of messages and its replies, even if it's a reply or it's in Tasks Topic of one user
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/messages/cascadeforce/topic/subTopic?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name
```

see https://github.com/ovh/tat#parameters for all parameters

## Create a task from a message
Add a message to topic: `/Private/username/Tasks`.

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "task" }'\
	https://<tatHostname>:<tatPort>/message/Private/username/Tasks
```

## Remove a message from tasks
Remove a message from the topic: /Private/username/Tasks

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "untask" }'\
	https://<tatHostname>:<tatPort>/message/Private/username/Tasks
```

## Vote UP a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "voteup" }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Remove a Vote UP from a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "unvoteup" }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Vote Down a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "votedown" }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Remove Vote Down from a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "unvotedown" }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

## Getting Messages List
```bash
curl -XGET https://<tatHostname>:<tatPort>/messages/<topic>?skip=<skip>&limit=<limit> | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/messages/<topic>?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name | python -m json.tool
```

Getting messages on one Public Topic (Read Only):

```bash
curl -XGET https://<tatHostname>:<tatPort>/read/<topic>?skip=<skip>&limit=<limit> | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/read/<topic>?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name | python -m json.tool
```

### Parameters

* `topic`: /yourTopic/subTopic
* `skip`: Skip skips over the n initial documents from the query results
* `limit`: Limit restricts the maximum number of documents retrieved
* `text`: your text
* `idMessage`: message Id
* `inReplyOfID`: message Id replied
* `inReplyOfIDRoot`: message Id root replied
* `allIDMessage`: message Id OR message Id replied OR message Id root replied
* `dateMinCreation`: filter result on dateCreation, timestamp Unix format
* `dateMaxCreation`: filter result on dateCreation, timestamp Unix Format
* `dateMinUpdate`: filter result on dateUpdate, timestamp Unix format
* `dateMaxUpdate`: filter result on dateUpdate, timestamp Unix Format
* `limitMinNbVotesUP`: filter result on nbVotesUP
* `limitMaxNbVotesUP`: filter result on nbVotesUP
* `limitMinNbVotesDown`: filter result on nbVotesDown
* `limitMaxNbVotesDown`: filter result on nbVotesDown
* `label`: Search by label: could be labelA,labelB
* `andLabel`: Search by label (and): could be labelA,labelB for select messages with labelA AND labelB
* `notLabel`: Search by label (exclude): could be labelA,labelB for select messages without labelA OR labelB
* `tag`: tagA,tagB
* `andTag`: tagA,tagB
* `notTag`: tagA,tagB
* `username`: usernameA,usernameB
* `treeView`: onetree or fulltree. "onetree": replies are under root message. "fulltree": replies are under their parent. Default: no tree
* `limitMinNbReplies`: in onetree mode, filter root messages with more or equals minNbReplies
* `limitMaxNbReplies`: in onetree mode, filter root messages with min or equals maxNbReplies
* `onlyMsgRoot`: restricts to root message only (inReplyOfIDRoot empty). If treeView is used, limit search criteria to root message, replies are still given, independently of search criteria.
* `onlyCount`: only count messages, without retrieve msg. limit, skip and treeview criterias are ignored

### Examples
```bash
curl -XGET https://<tatHostname>:<tatPort>/messages/topicA?skip=0&limit=100 | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/messages/topicA/subTopic?skip=0&limit=100&dateMinCreation=1405544146&dateMaxCreation=1405544146 | python -m json.tool
```
