import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [messages, setMessages] = useState([])
  const [isConnected, setIsConnected] = useState(false)

  useEffect(() => {
    const eventSource = new EventSource('http://localhost:8080/events')
    
    eventSource.onopen = () => {
      setIsConnected(true)
      console.log('SSE connection opened')
    }

    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data)
      setMessages(prev => [...prev, data])
    }

    eventSource.onerror = (error) => {
      setIsConnected(false)
      console.error('SSE error:', error)
    }

    return () => {
      eventSource.close()
    }
  }, [])

  const sendMessage = async () => {
    try {
      await fetch('http://localhost:8080/trigger', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ action: 'ping' })
      })
    } catch (error) {
      console.error('Error sending message:', error)
    }
  }

  const clearMessages = () => {
    setMessages([])
  }

  return (
    <div className="App">
      <header className="App-header">
        <h1>SSE Demo</h1>
        <div className="status">
          Connection: {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
        </div>
        
        <div className="controls">
          <button onClick={sendMessage}>Send Ping</button>
          <button onClick={clearMessages}>Clear Messages</button>
        </div>

        <div className="messages">
          <h3>Messages ({messages.length})</h3>
          <div className="message-list">
            {messages.map((msg, index) => (
              <div key={index} className="message">
                <span className="timestamp">{msg.timestamp}</span>
                <span className="content">{msg.message}</span>
                {msg.data && <span className="data">Data: {JSON.stringify(msg.data)}</span>}
              </div>
            ))}
          </div>
        </div>
      </header>
    </div>
  )
}

export default App
