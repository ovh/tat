[![Build Status](https://travis-ci.org/ovh/tat.svg?branch=master)](https://travis-ci.org/ovh/tat)
[![GoDoc](https://godoc.org/github.com/ovh/tat?status.svg)](https://godoc.org/github.com/ovh/tat)
[![Go Report Card](https://goreportcard.com/badge/ovh/tat)](https://goreportcard.com/report/ovh/tat)

# Tat Engine

<img align="right" src="https://raw.githubusercontent.com/ovh/tat/master/tat.png">

Tat, aka Text And Tags, is a communication tool - Human & Robot all together.

This is a pre-alpha version, but already known to be used in production.

Some use cases:
* Viewing Pull Requests, Build And Deployment in one place
* Alerting & Monitoring Overview
* Agile view as simple as a whiteboard with post-it
* Team Communication & Reporting facilities
* ...

Tat Engine exposes only an HTTP REST API.
You can manipulate this API with Tat Command Line Interface, aka **tatcli**, see
https://github.com/ovh/tatcli.

A **WebUI** is also available, see https://github.com/ovh/tatwebui.

Tat Engine:
* Uses MongoDB as backend
* Is fully stateless, scale as you want
* Is the central Hub of Tat microservices ecosystem

## General Specifications

* Topic
 * Contains 0 or n messages
 * Administrator(s) of Topic can create Topic inside it
* Message
 * Consists of text, tags and labels
 * Can not be deleted or modified (by default)
 * Is limited in characters (topic setting)
 * Is always attached to one topic
* Tag
 * Within the message content
 * Can not be added after message creation (by default)
* Label
 * Can be added or removed freely
 * Have a color
* Group
 * Managed by an administrator(s): adding or removing users from the group
 * Without prior authorization, a group or user has no access to topics
 * A group or a user can be read-only or read-write on a topic
* Task
 * A *task* is a message that is both in the topic task of a user and in the original topic
* Administrator(s)
 * Tat Administrator: all configuration access
 * On Group(s): can add/remove member(s)
 * On Topic(s): can create Topic inside it, update parameters

### Some rules and rules exception:
* Deleting a message is possible in the private topics, or can be granted on other topic
* Modification of a message is possible in private topics, or can be granted on other topic
* The default length of a message is 140 characters, this limit can be modified by topic
* A date creation of a message can be explicitly set by a `system user`
* message.dateCreation and message.dateUpdate are in timestamp format, ex:
 * 1436912447: 1436912447 seconds
 * 1436912447.345678: 1436912447 seconds and 345678 milliseconds

### FAQ:
*What about attachment (sound, image, etc...) ?*
Tat Engine stores only *text*. Use other application, like Plik (https://github.com/root-gg/plik)
to upload file and store URL on Tat. This workflow should be done by UI.


# Usage of Tat Engine API
## Message
### Store new message
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

### Store some messages
```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '[{ "text": "text message A" },{ "text": "text message B", "labels": [{"text": "labelA", "color": "#eeeeee"}] }]' \
	https://<tatHostname>:<tatPort>/messages/topic/sub-topic
```

### Reply to a message
```bash
curl -XPOST \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "text": "text", "idReference": "9797q87KJhqsfO7Usdqd", "action": "reply"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Like a message
```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "like"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Unlike a message
```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "unlike"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Add a label to a message
*option* is the background color of the label.

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "label", "text": "myLabel", "option": "rgba(143,199,148,0.61)"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Remove a label from a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "unlabel", "text": "myLabel"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Remove all labels and add new ones

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "relabel", "labels": [{"text": "labelA", "color": "#eeeeee"}, {"text": "labelB", "color": "#ffffff"}]}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

Return HTTP 201 if OK

### Remove some labels and add new ones

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "relabel", "labels": [{"text": "labelA", "color": "#eeeeee"}, {"text": "labelB", "color": "#ffffff"}], "options": ["labelAToRemove", "labelAToRemove"] }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Update a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "update", "text": "my New Mesage updated"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Concat a message : adding additional text to one message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "concat", "text": " additional text"}'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Move a message to another topic

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "move", "option": "/newTopic/subNewTopic"}'\
	https://<tatHostname>:<tatPort>/message/oldTOpic/oldSubTopic
```

### Delete a message
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/message/nocascade/9797q87KJhqsfO7Usdqd/topic/subTopic
```

### Delete a message and its replies
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/message/cascade/9797q87KJhqsfO7Usdqd/topic/subTopic
```

### Delete a message and its replies, even if it's in Tasks Topic of one user
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/message/cascadeforce/9797q87KJhqsfO7Usdqd/topic/subTopic
```

### Delete a list of messages
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/messages/nocascade/topic/subTopic?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name
```

see https://github.com/ovh/tat#parameters for all parameters

### Delete a list of messages and its replies
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/messages/cascade/topic/subTopic?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name
```

see https://github.com/ovh/tat#parameters for all parameters

### Delete a list of messages and its replies, even if it's a reply or it's in Tasks Topic of one user
```bash
curl -XDELETE \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/messages/cascadeforce/topic/subTopic?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name
```

see https://github.com/ovh/tat#parameters for all parameters

### Create a task from a message
Add a message to topic: `/Private/username/Tasks`.

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "task" }'\
	https://<tatHostname>:<tatPort>/message/Private/username/Tasks
```

### Remove a message from tasks
Remove a message from the topic: /Private/username/Tasks

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "untask" }'\
	https://<tatHostname>:<tatPort>/message/Private/username/Tasks
```

### Vote UP a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "voteup" }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Remove a Vote UP from a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "unvoteup" }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Vote Down a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "votedown" }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Remove Vote Down from a message

```bash
curl -XPUT \
    -H 'Content-Type: application/json' \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "idReference": "9797q87KJhqsfO7Usdqd", "action": "unvotedown" }'\
	https://<tatHostname>:<tatPort>/message/topic/sub-topic
```

### Getting Messages List
```bash
curl -XGET https://<tatHostname>:<tatPort>/messages/<topic>?skip=<skip>&limit=<limit> | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/messages/<topic>?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name | python -m json.tool
```

Getting messages on one Public Topic (Read Only):

```bash
curl -XGET https://<tatHostname>:<tatPort>/read/<topic>?skip=<skip>&limit=<limit> | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/read/<topic>?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name | python -m json.tool
```

#### Parameters

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

#### Examples
```bash
curl -XGET https://<tatHostname>:<tatPort>/messages/topicA?skip=0&limit=100 | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/messages/topicA/subTopic?skip=0&limit=100&dateMinCreation=1405544146&dateMaxCreation=1405544146 | python -m json.tool
```

### Convert a user to a system user
Only for Tat Admin: convert a `normal user` to a `system user`.
A system user must have a username starting with `tat.system`.
Remove email and set user attribute IsSystem to true.
This action returns a new password for this user.
Warning: it is an irreversible action.

Flag `canWriteNotifications` allows (or not if false) the `system user` to write inside private topics of user `/Private/username/Notifications`

Flag `canListUsersAsAdmin` allows this `system user` to view all user's fields (email, etc...)

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: userAdmin" \
    -H "Tat_password: passwordAdmin" \
    -d '{ "username": "usernameToConvert", "canWriteNotifications": "true", "canListUsersAsAdmin": "true" }' \
    https://<tatHostname>:<tatPort>/user/convert
```

### Update flags on system user
Only for Tat Admin.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: userAdmin" \
    -H "Tat_password: passwordAdmin" \
    -d '{ "username": "userSystem", "canWriteNotifications": "true", "canListUsersAsAdmin": "true" }' \
    https://<tatHostname>:<tatPort>/user/updatesystem
```

### Reset a password for system user
Only for Tat Admin.
A `system user` must have a username starting with `tat.system`.
This action returns a new password for this user.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: userAdmin" \
    -H "Tat_password: passwordAdmin" \
    -d '{ "username": "userameSystemToReset" }' \
    https://<tatHostname>:<tatPort>/user/resetsystem
```


### Grant a user to an admin user
Only for Tat Admin: convert a `normal user` to an `admin user`.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: userAdmin" \
    -H "Tat_password: passwordAdmin" \
    -d '{ "username": "usernameToGrant" }' \
    https://<tatHostname>:<tatPort>/user/setadmin
```

### Rename a username
Only for Tat Admin: rename the username of a user. This action updates all Private topics of the user.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: userAdmin" \
    -H "Tat_password: passwordAdmin" \
    -d '{ "username": "usernameToRename", "newUsername": "NewUsername" }' \
    https://<tatHostname>:<tatPort>/user/rename
```

### Update fullname or email
Only for Tat Admin: update fullname and email of a user.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: userAdmin" \
    -H "Tat_password: passwordAdmin" \
    -d '{ "username": "usernameToRename", "newFullname": "NewFullname", "newEmail": "NewEmail" }' \
    https://<tatHostname>:<tatPort>/user/update
```

### Archive a user
Only for Tat Admin

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: userAdmin" \
    -H "Tat_password: passwordAdmin" \
    -d '{ "username": "usernameToRename" }' \
    https://<tatHostname>:<tatPort>/user/archive
```

### Check Private Topics and Default Group on one user
Only for Tat Admin

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: userAdmin" \
    -H "Tat_password: passwordAdmin" \
    -d '{ "username": "usernameToRename",  "fixPrivateTopics": true, "fixDefaultGroup": true }' \
    https://<tatHostname>:<tatPort>/user/check
```

## Presence
### Add presence
Status could be: `online`, `offline`, `busy`.

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	-d '{ "status": "online" }' \
	https://<tatHostname>:<tatPort>/presenceget/topic/sub-topic
```

### Getting Presences
```bash
curl -XGET https://<tatHostname>:<tatPort>/presences/<topic>?skip=<skip>&limit=<limit> | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/presences/<topic>?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name | python -m json.tool
```

### Parameters

* `topic:` /yourTopic/subTopic
* `skip`: Skip skips over the n initial presences from the query results
* `limit`: Limit restricts the maximum number of presences retrieved
* `status`: status: `online`, `offline`, `busy`
* `dateMinPresence`: filter result on datePresence, timestamp Unix format
* `dateMaxPresence`: filter result on datePresence, timestamp Unix Format
* `username`: username to search


#### Examples
```bash
curl -XGET https://<tatHostname>:<tatPort>/presences/topicA?skip=0&limit=100 | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/presences/topicA/subTopic?skip=0&limit=100&dateMinPresence=1405544146&dateMaxPresence=1405544146 | python -m json.tool
```

### Delete presence
Admin can delete presences a another user on one topic.
Users can delete their own presence.

```bash
curl -XDELETE \
    -H "Content-Type: application/json" \
    -H "Tat_username: username" \
    -H "Tat_password: passwordOfUser" \
	https://<tatHostname>:<tatPort>/presences/topic/sub-topic
```

## User
### Tat Password
It's a generated password by Tat, allowing username to communicate with Tat.
User creates an account, a mail is send to verify account and user has to go on a Tat URL to validate account and get password.
Password is encrypted in Tat Database (sha512 Sum).

First user created is an administrator.

### Create a User
Return a mail to user, with instruction to validate his account.

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -d '{"username": "userA", "fullname": "User AA", "email": "usera@foo.net", "callback": " Click on:scheme://:host:port/user/verify/:token to validate your account"}' \
    https://<tatHostname>:<tatPort>/user
```

Callback is a string sent by mail, indicating to the user how to validate his account.
Available fields (automatically filled by Tat ):

```
:scheme -> http of https
:host -> ip or hostname of Tat Engine
:port -> port of Tat Engine
:username -> username
:token -> tokenVerify of user
```


### Verify a User
```bash
curl -XGET \
    https://<tatHostname>:<tatPort>/user/verify/yourUsername/tokenVerifyReceivedByMail
```
This url can be called only once per password and expired 30 minutes after querying create user with POST on `/user`

### Ask for reset a password
Returns: tokenVerify by email

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -d '{"username": "userA", "email": "usera@foo.net"}' \
    https://<tatHostname>:<tatPort>/user/reset
```

### Get information about current User
```bash
curl -XGET \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA" \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me
```


### Get contacts

Retrieves contacts presences since n seconds

Example since 15 seconds :

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA" \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/contacts/15
```

### Add a contact
```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA" \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/contact/username
```

### Remove a contact
```bash
curl -XDELETE \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA" \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/contacts/username
```


### Add a favorite topic
```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/topics/myTopic/sub-topic
```

### Remove a favorite topic
```bash
curl -XDELETE \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA" \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/topics/myTopic/sub-topic
```

### Enable notifications on one topic
```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/enable/notifications/topics/myTopic/sub-topic
```

### Disable notifications on one topic
```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA" \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/disable/notifications/topics/myTopic/sub-topic
```

### Enable notifications on all topics
```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/enable/notifications/alltopics
```

### Disable notifications on all topics, except /Private/*
```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA" \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/disable/notifications/alltopics
```

### Add a favorite tag
```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA" \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/tags/myTag
```

### Remove a favorite tag
```bash
curl -XDELETE \
    -H "Content-Type: application/json" \
    -H "Tat_username: userA" \
    -H "Tat_password: password" \
    https://<tatHostname>:<tatPort>/user/me/tags/myTag
```

### Getting Users List
```bash
curl -XGET https://<tatHostname>:<tatPort>/users?skip=<skip>&limit=<limit> | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/users?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name | python -m json.tool
```

Users list with groups (admin only)
```bash
curl -XGET https://<tatHostname>:<tatPort>/users?skip=<skip>&limit=<limit>&withGroups=true
```

#### Parameters

* skip: Skip skips over the n initial documents from the query results
* limit: Limit restricts the maximum number of documents retrieved
* username: Username
* fullname: Fullname
* dateMinCreation: filter result on dateCreation, timestamp Unix format
* dateMaxCreation: filter result on dateCreation, timestamp Unix Format


#### Example
```bash
curl -XGET https://<tatHostname>:<tatPort>/users?skip=0&limit=100 | python -m json.tool
```

## Group
### Create a group

Only for Tat Admin

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"name": "groupName", "description": "Group Description"}' \
    https://<tatHostname>:<tatPort>/group
```

### Update a group

Only for Tat Admin

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"newName": "groupName", "newDescription": "Group Description"}' \
    https://<tatHostname>:<tatPort>/group/<groupName>
```

### Getting groups List

```bash
curl -XGET https://<tatHostname>:<tatPort>/groups?skip=<skip>&limit=<limit> | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/groups?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name | python -m json.tool
```

#### Parameters

* skip: Skip skips over the n initial documents from the query results
* limit: Limit restricts the maximum number of documents retrieved
* idGroup: Id Group
* name: Name of group
* description: Description of group
* dateMinCreation: filter result on dateCreation, timestamp Unix format
* dateMaxCreation: filter result on dateCreation, timestamp Unix Format

### Delete a group

Only for Tat Admin

```bash
curl -XDELETE \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/group/<groupName>
```

### Add a user to a group
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"groupname": "groupName", "username": "usernameToAdd"}' \
    https://<tatHostname>:<tatPort>/group/add/user
```

### Delete a user from a group
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"groupname": "groupName", "username": "usernameToAdd"}' \
    https://<tatHostname>:<tatPort>/group/remove/user
```


### Add an admin user to a group
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"groupname": "groupName", "username": "usernameToAdd"}' \
    https://<tatHostname>:<tatPort>/group/add/adminuser
```

### Delete an admin user from a group
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"groupname": "groupName", "username": "usernameToAdd"}' \
    https://<tatHostname>:<tatPort>/group/remove/adminuser
```


## Topic
### Create a Topic

Rules:

* User can create a root topic if he is a Tat Admin.
* User can create topics under `/Private/username/`
* User can create topics if he is an admin on the Parent Topic or belong to an admin group on the Parent topic.
Example:  Create /AAA/BBB: Parent Topic is /AAA

```bash
curl -XPOST \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "description": "Topic Description"}' \
    https://<tatHostname>:<tatPort>/topic
```

### Delete a topic
```bash
curl -XDELETE \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/topic/subtopic
```

### Truncate a topic

Only for Tat Admin and administrators on topic.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA"}' \
    https://<tatHostname>:<tatPort>/topic/truncate
```

### Compute tags on a topic

Only for Tat Admin and administrators on topic.

Set "tags" attribute on topic, with an array of all tags used in this topic.
One entry in "tags" attribute per text of tag.

Topic's tags are showed with :
GET https://<tatHostname>:<tatPort>/topic/topicName

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA"}' \
    https://<tatHostname>:<tatPort>/topic/compute/tags
```

Example of usage of tags attribute: autocompletion of tag on UI when written new message on a topic

### Compute labels on a topic

Only for Tat Admin and administrators on topic.

Set "labels" attribute on topic, with an array of all labels used in this topic.
One entry in "labels" attribute per text & color of label.

Topic's labels are showed with :
GET https://<tatHostname>:<tatPort>/topic/topicName

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA"}' \
    https://<tatHostname>:<tatPort>/topic/compute/labels
```

Example of usage of labels attribute: label autocompletion on UI when adding new label

### Compute tags on all topics

Only for Tat Admin.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/topics/compute/tags
```

### Compute labels on all topics

Only for Tat Admin.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/topics/compute/labels
```

### Set a param on all topics

Only for Tat Admin and for attributes isAutoComputeTags and isAutoComputeLabels.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"paramName":"isAutoComputeLabels","paramValue":"false"}' \
    https://<tatHostname>:<tatPort>/topics/param
```


### Truncate cached tags on a topic

Only for Tat Admin and administrators on topic.

Truncate "tags" attribute on topic.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA"}' \
    https://<tatHostname>:<tatPort>/topic/tags/truncate
```

### Truncate cached labels on a topic

Only for Tat Admin and administrators on topic.

Truncate "labels" attribute on topic.

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA"}' \
    https://<tatHostname>:<tatPort>/topic/labels/truncate
```

### Getting one Topic
```bash
curl -XGET https://<tatHostname>:<tatPort>/topic/topicName | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/topic/topicName/subTopic | python -m json.tool
```

### Getting Topics List
```bash
curl -XGET https://<tatHostname>:<tatPort>/topics?skip=<skip>&limit=<limit> | python -m json.tool
curl -XGET https://<tatHostname>:<tatPort>/topics?skip=<skip>&limit=<limit>&argName=valName&arg2Name=val2Name | python -m json.tool
```

#### Parameters
* skip: Skip skips over the n initial documents from the query results
* limit: Limit restricts the maximum number of documents retrieved
* topic: Topic name, example: /topicA
* topicPath: Topic start path, example: /topicA will return /topicA/subA, /topicA/subB
* idTopic: id of topic
* description: description of topic
* dateMinCreation: filter result on dateCreation, timestamp Unix format
* dateMaxCreation: filter result on dateCreation, timestamp Unix Format
* getNbMsgUnread: if true, add new array to return, topicsMsgUnread with topic:flag. flag can be -1 if unknown, 0 or 1 if there is one or more messages unread
* onlyFavorites: if true, return only favorites topics, except /Private/*. All privates topics are returned.
* getForTatAdmin: if true, and requester is a Tat Admin, returns all topics (except /Private/*) without checking user access


#### Example
```bash
curl -XGET https://<tatHostname>:<tatPort>/topics?skip=0&limit=100 | python -m json.tool
```

### Add a parameter to a topic

For admin of topic or on `/Private/username/*`

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "key": "keyOfParameter", "value": "valueOfParameter", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/add/parameter
```

### Remove a parameter to a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "key": "keyOfParameter", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/remove/parameter
```

### Add a read only user to a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "username": "usernameToAdd", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/add/rouser
```

### Add a read write user to a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "username": "usernameToAdd", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/add/rwuser
```

### Add an admin user to a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "username": "usernameToAdd", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/add/adminuser
```

### Delete a read only user from a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "username": "usernameToRemove", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/remove/rouser
```

### Delete a read write user from a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "username": "usernameToRemove", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/remove/wuser
```

### Delete an admin user from a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "username": "usernameToRemove", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/remove/adminuser
```

### Add a read only group to a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "groupname": "groupnameToAdd", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/add/rogroup
```

### Add a read write group to a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "groupname": "groupnameToAdd", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/add/rwgroup
```

### Add an admin group to a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "groupname": "groupnameToAdd", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/add/admingroup
```


### Delete a read only group from a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "groupname": "groupnameToRemove", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/remove/rogroup
```

### Delete a read write group from a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "groupname": "groupnameToRemove", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/remove/rwgroup
```

### Delete an admin group from a topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "groupname": "groupnameToRemove", "recursive": "false"}' \
    https://<tatHostname>:<tatPort>/topic/remove/rwgroup
```


### Update param on one topic: admin or admin on topic
```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic": "/topicA", "recursive": "false", "maxlength": 140, "canForceDate": false, "canUpdateMsg": false, "canDeleteMsg": false, "canUpdateAllMsg": false, "canDeleteAllMsg": false, "adminCanUpdateAllMsg": false, "adminCanDeleteAllMsg": false}' \
    https://<tatHostname>:<tatPort>/topic/param
```

Parameters key is optional.

Example with key parameters :

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    -d '{"topic":"/Internal/Alerts","recursive":false,"maxlength":300,"canForceDate":false,"canUpdateMsg":false,"canDeleteMsg":true,"canUpdateAllMsg":false,"canDeleteAllMsg":false,"adminCanUpdateAllMsg":false,"adminCanDeleteAllMsg":false,"parameters":[{"key":"agileview","value":"qsdf#qsdf"},{"key":"tatwebui.view.default","value":"standardview-list"},{"key":"tatwebui.view.forced","value":""}]}' \
    https://<tatHostname>:<tatPort>/topic/param
```

## System
### Version

```bash
curl -XGET https://<tatHostname>:<tatPort>/version
```


## Stats

For Tat admin only.

### Count

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/stats/count
```

### Instance

Info about current instance of engine

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/stats/instance
```

### Distribution

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/stats/distribution
```

### DB Stats

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/stats/db/stats
```

### DB ServerStatus

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/stats/db/serverStatus
```

### DB Replica Set Status

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/stats/db/replSetGetStatus
```

### DB Replica Set Config

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/stats/db/replSetGetConfig
```



### DB Stats of each collections

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/stats/db/collections
```

### DB Stats slowest Queries

```bash
curl -XPUT \
    -H "Content-Type: application/json" \
    -H "Tat_username: admin" \
    -H "Tat_password: passwordAdmin" \
    https://<tatHostname>:<tatPort>/stats/db/slowestQueries
```

## System
### Capabilities

Return `websocket-enabled` and `username-from-email` parameters. See Tat Flags below.
```bash
curl -XGET https://<tatHostname>:<tatPort>/capabilities
```

### Flush Cache
```bash
curl -XGET https://<tatHostname>:<tatPort>/system/cache/clean
```

### Cache Info
```bash
curl -XGET https://<tatHostname>:<tatPort>/system/cache/info
```

# RUN

## Tat Flags Options

```
Flags:
      --allowed-domains string                      Users have to use theses emails domains. Empty: no-restriction. Ex: --allowed-domains=domainA.org,domainA.com
      --db-addr string                              Address of the mongodb server (default "127.0.0.1:27017")
      --db-password string                          Password to authenticate with the mongodb server. If "false", db-password is not used
      --db-rs-tags string                           Link hostname with tag on mongodb replica set - Optional: hostnameA:tagName:value,hostnameB:tagName:value. If "false", db-rs-tags is not used
      --db-socket-timeout int                       Session DB Socket Timeout in seconds (default 60)
      --db-user string                              User to authenticate with the mongodb server. If "false", db-user is not used
      --default-domain string                       Default domains for mail for trusted username
      --default-group string                        Default Group for new user
      --exposed-host string                         Tat Engine Hostname exposed to client (default "localhost")
      --exposed-path string                         Tat Engine Path exposed to client, ex: host:port/tat/engine /tat/engine is exposed path
      --exposed-port string                         Tat Engine Port exposed to client (default "8080")
      --exposed-scheme string                       Tat URI Scheme http or https exposed to client (default "http")
      --header-trust-username string                Header Trust Username: for example, if X-Remote-User and X-Remote-User received in header -> auto accept user without testing tat_password. Use it with precaution
      --listen-port string                          Tat Engine Listen Port (default "8080")
      --no-smtp                                     No SMTP mode
      --production                                  Production mode
      --read-timeout int                            Read Timeout in seconds (default 50)
      --redis-hosts string                          Optional - Used for Cache - Redis hosts (comma separated for cluster)
      --redis-password string                       Optional - Used for Cache - Redis password
      --smtp-from string                            SMTP From
      --smtp-host string                            SMTP Host
      --smtp-password string                        SMTP Password
      --smtp-port string                            SMTP Port
      --smtp-tls                                    SMTP TLS
      --smtp-user string                            SMTP Username
      --tat-log-level string                        Tat Log Level: debug, info or warn
      --trusted-usernames-emails-fullnames string   Tuples trusted username / email / fullname. Example: username:email:Firstname1_Fullname1,username2:email2:Firstname2_Fullname2
      --username-from-email                         Username are extracted from first part of email. first.lastame@domainA.org -> username: first.lastname
      --websocket-enabled                           Enable or not websockets on this instance
      --write-timeout int                           Write Timeout in seconds (default 50)

```

## Run with Docker
```bash
git clone https://github.com/ovh/tat.git && cd tat
docker run --name tat-mongo -d -v /home/yourhome/data:/data/db mongo
docker build -t tat .
docker run -it --rm --name tat-instance1 --link tat-mongo:mongodb \
  -e TAT_DB_ADDR=mongodb:27017 \
  -e TAT_SMTP_HOST=yourSMTPHost \
  -e TAT_SMTP_PORT=yourSMTPPort \
  -e TAT_SMTP_USER=yourSMTPUser \
  -e TAT_SMTP_TLS=false \
  -e TAT_SMTP_PASSWORD=yourSMTPPassword \
  -e TAT_EXPOSED_HOST=<hostnameOfTat> \
  -e TAT_EXPOSED_PORT=8080 \
  -e TAT_EXPOSED_SCHEME=http \
  -e production=true -p 8080:8080 tat
```

## Dev RUN

```bash
cd cd $GOPATH/src/github.com/ovh/tat/api && go build && ./api --no-smtp=true --help
```

If you want to create a user with tatcli:

```bash
tatcli --url="https://localhost:8080" user add yourUsername firstname.lastname@ovh.net Firstname Lastname
```


### Environment

* TAT_LISTEN_PORT
* TAT_DB_ADDR
* TAT_DB_USER
* TAT_DB_PASSWORD

Example:
```bash
export TAT_LISTEN_PORT=8181 && ./tat
```
is same than
```bash
./tat --listen-port="8181"
```


# SDK

## Documentation

See https://godoc.org/github.com/ovh/tat#Client

## Example - Minimal

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ovh/tat"
)

// taturl, username / password of tat engine
var (
	taturl   string
	username string
	password string
)

/*
Usage:
 go get -u github.com/ovh/tat
 build && ./mycli-minimal -url=http://url-tat-engine -username=<tatUsername> -password=<tatPassword> /Internal/your/topic your message
*/

func main() {
	flag.StringVar(&taturl, "url", "", "URL of Tat Engine")
	flag.StringVar(&username, "username", "", "tat username")
	flag.StringVar(&password, "password", "", "tat password")
	flag.Parse()

	client, err := tat.NewClient(tat.Options{
		URL:      taturl,
		Username: username,
		Password: password,
		Referer:  "mycli-minimal.v0",
	})

	if err != nil {
		fmt.Printf("Error while create new Tat Client: %s\n", err)
		os.Exit(1)
	}

	args := flag.Args()
	text := strings.Join(args[1:], " ")
	topic := args[0]
	m := tat.MessageJSON{
		Text:  text,
		Topic: topic,
	}

	fmt.Printf("Send on topic %s this message: %s\n", topic, text)

	msgCreated, err := client.MessageAdd(m)
	if err != nil {
		fmt.Printf("Error:%s\n", err)
		os.Exit(1)
	}
	fmt.Printf("ID Message Created: %s\n", msgCreated.Message.ID)
}

```

## Example - Full, with viper, cobra and tatcli config file

```go
package main

/*
Usage:
 go build && ./mycli-full demo /YouTopic/subTopic your message

with a config file:
 go build && ./mycli-full --configFile $HOME/.tatcli/config.local.json demo /YouTopic/subTopic your message

You should split this file into many files.
See https://github.com/ovh/tatcli for CLI with many subcommands
*/

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:  "mycli-full",
	Long: `SDK Use Demo`,
}

// URL of tat engine, tat username and tat password
var (
	home     = os.Getenv("HOME")
	taturl   string
	username string
	password string

	// ConfigFile is $HOME/.tatcli/config.json per default
	// contains user, password and url of tat
	configFile string
)

func main() {
	addCommands()

	rootCmd.PersistentFlags().StringVarP(&taturl, "url", "", "", "URL Tat Engine, facultative if you have a "+home+"/.tatcli/config.json file")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "username, facultative if you have a "+home+"/.tatcli/config.json file")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "password, facultative if you have a "+home+"/.tatcli/config.json file")
	rootCmd.PersistentFlags().StringVarP(&configFile, "configFile", "c", home+"/.tatcli/config.json", "configuration file, default is "+home+"/.tatcli/config.json")

	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))

	log.SetLevel(log.DebugLevel)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

//AddCommands adds child commands to the root command rootCmd.
func addCommands() {
	rootCmd.AddCommand(cmdDemo)
}

var cmdDemo = &cobra.Command{
	Use:   "demo <topic> <msg>",
	Short: "Demo Post Msg",
	Run: func(cmd *cobra.Command, args []string) {
		create(args[0], args[1])
	},
}

// create creates a message in specified topic
func create(topic, message string) {
	readConfig()
	m := tat.MessageJSON{Text: message, Topic: topic}
	msgCreated, err := getClient().MessageAdd(m)
	if err != nil {
		log.Errorf("Error:%s", err)
		return
	}
	log.Debugf("ID Message Created: %d", msgCreated.Message.ID)
}

func getClient() *tat.Client {
	tc, err := tat.NewClient(tat.Options{
		URL:      viper.GetString("url"),
		Username: viper.GetString("username"),
		Password: viper.GetString("password"),
		Referer:  "mycli.v0",
	})

	if err != nil {
		log.Fatalf("Error while create new Tat Client: %s", err)
	}

	tat.DebugLogFunc = log.Debugf
	return tc
}

// readConfig reads config in .tatcli/config per default
func readConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
		viper.ReadInConfig() // Find and read the config file
	}
}

```


# Hacking

Best with go >= 1.7.

```bash
mkdir -p $GOPATH/src/github.com/ovh
cd $GOPATH/src/github.com/ovh
git clone git@github.com:ovh/tat.git
cd $GOPATH/src/github.com/ovh/tat/api
go build
```

You've developed a new cool feature? Fixed an annoying bug? We'd be happy
to hear from you! Make sure to read [CONTRIBUTING.md](./CONTRIBUTING.md) before.
