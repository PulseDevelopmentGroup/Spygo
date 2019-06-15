class MessageBroker {
  constructor() {
    this.subs = {};
  }
  
  handleMessage(message) {
    const msg = this.parseMessage(message);
    console.log(msg);
    
    this.subs[msg.type].forEach((cb) => {
      cb(msg);
    })
  }
  
  subscribe(event, cb) {
    if (!this.subs[event]) {
      this.subs[event] = [];
    }
    
    this.subs[event].push(cb);
  }
  
  unsubscribe(event, cb) {
    if (!this.subs[event]) {
      return;
    }
    
    delete this.subs[event];
  }
  
  parseMessage(e) {
    const parsedMessage = JSON.parse(e.data);
    try {
      if (!parsedMessage.error) {
        return {
          type: parsedMessage.kind,
          data: JSON.parse(parsedMessage.data),
        }
      } else {
        return {
          type: parsedMessage.kind,
          data: JSON.parse(parsedMessage.data),
          error: JSON.parse(parsedMessage.data)
        }
      }
    } catch (err) {
      console.log('Error occured while parsing socket message:', err);
    }
  }
  
}

export default new MessageBroker();