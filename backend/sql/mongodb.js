/*
 Navicat Premium Dump Script

 Source Server         : A6000
 Source Server Type    : MongoDB
 Source Server Version : 80301 (8.3.1)
 Source Host           : 10.16.15.202:27017
 Source Schema         : wetalk

 Target Server Type    : MongoDB
 Target Server Version : 80301 (8.3.1)
 File Encoding         : 65001

 Date: 17/05/2026 21:03:19
*/


// ----------------------------
// Collection structure for chat_deletions
// ----------------------------
db.getCollection("chat_deletions").drop();
db.createCollection("chat_deletions");
db.getCollection("chat_deletions").createIndex({
    user_id: Int32("1"),
    chat_id: Int32("1")
}, {
    name: "user_id_1_chat_id_1",
    unique: true
});

// ----------------------------
// Collection structure for counters
// ----------------------------
db.getCollection("counters").drop();
db.createCollection("counters");

// ----------------------------
// Collection structure for messages
// ----------------------------
db.getCollection("messages").drop();
db.createCollection("messages");
db.getCollection("messages").createIndex({
    chat_id: Int32("1"),
    created_at: Int32("-1")
}, {
    name: "chat_id_1_created_at_-1"
});
db.getCollection("messages").createIndex({
    chat_id: Int32("1"),
    sender_id: Int32("1"),
    status: Int32("1")
}, {
    name: "chat_id_1_sender_id_1_status_1"
});
db.getCollection("messages").createIndex({
    msg_id: Int32("1")
}, {
    name: "msg_id_1",
    unique: true
});
