"use strict";

//getMemberIDs returns user ids of all
//members of given channel
async function getMemberIDs(channelID, db) {
    const memberIDsArray = [];
    try {
      const qry = "SELECT MemberID FROM ChannelsJoinMembers C WHERE C.ChannelID = ?";
      const user = await db.query(qry, [channelID]);
      for (let i = 0; i < user.length; i++) {
        memberIDsArray.push(user[i]);
      }
      return { members: memberIDsArray, error: null };
    } catch (err) {
      return { members: [], error: err };
    }
}

/**
 * Expose public helper functions.
 */
module.exports = {
    getMemberIDs
}