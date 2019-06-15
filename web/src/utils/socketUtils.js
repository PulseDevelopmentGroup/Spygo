/**
 *  Create a payload object for transmission over socket
 * @param {string} kind - The type of message
 * @param {string} data - The payload of the message
 */
function packMessage(kind, data) {
  let payload;

  if (typeof data !== 'string') {
    try {
      payload = JSON.stringify(data);  
    } catch (e) {
      console.error('An error occured while trying to stringify the payload: ', e);
    }
  } else {
    payload = data
  }
  
  console.log(payload);
  
  return JSON.stringify({
    kind,
    data: payload
  });
}

export {
  packMessage
}