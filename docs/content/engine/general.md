---
weight: 1
toc: true
title: "General"
prev: "/engine"
next: "/engine/api-messages"

---


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

## Some rules and rules exception
* Deleting a message is possible in the private topics, or can be granted on other topic
* Modification of a message is possible in private topics, or can be granted on other topic
* The default length of a message is 140 characters, this limit can be modified by topic
* A date creation of a message can be explicitly set by a system user
* message.dateCreation and message.dateUpdate are in timestamp format, ex:
 * 1436912447: 1436912447 seconds
 * 1436912447.345678: 1436912447 seconds and 345678 milliseconds

## FAQ
*What about attachment (sound, image, etc...) ?*
Tat Engine stores only *text*. Use other application, like Plik (https://github.com/root-gg/plik)
to upload file and store URL on Tat. This workflow should be done by UI.
