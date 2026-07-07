import asyncio
import json
import random
import string
import sys
import requests
import websockets

BASE_URL = "http://localhost:8080/api/v1"
WS_URL = "ws://localhost:8080/api/v1/ws/connect"

def rand_string(length=8):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def register_or_login(email, password, first_name, last_name, phone, account_type):
    # Try registering
    reg_payload = {
        "first_name": first_name,
        "last_name": last_name,
        "email": email,
        "phone": phone,
        "password": password,
        "account_type": account_type
    }
    r = requests.post(f"{BASE_URL}/auth/register", json=reg_payload)
    if r.status_code == 201:
        print(f"Registered user: {email}")
    elif r.status_code == 409:
        print(f"User already exists (Conflict), proceeding to login: {email}")
    else:
        print(f"Registration failed for {email}: {r.status_code} {r.text}")

    # Login
    login_payload = {
        "email": email,
        "password": password
    }
    r = requests.post(f"{BASE_URL}/auth/login", json=login_payload)
    if r.status_code != 200:
        print(f"Login failed for {email}: {r.status_code} {r.text}")
        sys.exit(1)
    
    data = r.json()
    return data["access_token"], data["user"]["id"]

async def main():
    # Generate static test accounts
    email_a = "ws_test_user_a@homerent.zm"
    email_b = "ws_test_user_b@homerent.zm"
    password = "SecurePassword123!"

    print("--- 1. Authenticating Users ---")
    token_a, id_a = register_or_login(email_a, password, "Alice", "Test", "0971111111", "tenant")
    token_b, id_b = register_or_login(email_b, password, "Bob", "Test", "0972222222", "landlord")

    print(f"User A ID: {id_a}")
    print(f"User B ID: {id_b}")

    print("\n--- 2. Creating Conversation ---")
    headers_a = {"Authorization": f"Bearer {token_a}"}
    conv_payload = {
        "participant_id": id_b,
        "property_id": "prop_test_ws_123"
    }
    r = requests.post(f"{BASE_URL}/conversations", json=conv_payload, headers=headers_a)
    if r.status_code not in (200, 201):
        print(f"Failed to create conversation: {r.status_code} {r.text}")
        sys.exit(1)
    conv_data = r.json()
    print(f"Conversation Data: {conv_data}")
    conv_id = conv_data.get("ID") or conv_data.get("id")
    print(f"Conversation ID: {conv_id}")

    print("\n--- 3. Connecting User A & B to WebSocket ---")
    
    # Connect User A
    ws_a = await websockets.connect(
        WS_URL,
        extra_headers={"Authorization": f"Bearer {token_a}"}
    )
    # Connect User B
    ws_b = await websockets.connect(
        WS_URL,
        extra_headers={"Authorization": f"Bearer {token_b}"}
    )

    print("Connected both clients.")

    # Read welcome message for User A
    msg_a = await ws_a.recv()
    welcome_a = json.loads(msg_a)
    print(f"User A Welcome: {welcome_a}")
    assert welcome_a["event"] == "ping", "Expected ping event for welcome"

    # Read welcome message for User B
    msg_b = await ws_b.recv()
    welcome_b = json.loads(msg_b)
    print(f"User B Welcome: {welcome_b}")
    assert welcome_b["event"] == "ping", "Expected ping event for welcome"

    print("\n--- 4. User A sends a Message to User B ---")
    payload = {
        "event": "message",
        "payload": {
            "conversation_id": conv_id,
            "content": "Hello Bob, this is Alice!"
        }
    }
    await ws_a.send(json.dumps(payload))
    print("Alice sent message.")

    # User B should receive the message
    msg_b_recv = await ws_b.recv()
    event_b = json.loads(msg_b_recv)
    print(f"Bob received event: {event_b}")
    assert event_b["event"] == "message", "Expected message event"
    payload_b = event_b["payload"]
    assert (payload_b.get("content") or payload_b.get("Content")) == "Hello Bob, this is Alice!", "Content mismatch"

    # Alice (User A) should receive the echo message to confirm delivery
    msg_a_recv = await ws_a.recv()
    event_a = json.loads(msg_a_recv)
    print(f"Alice received echo event: {event_a}")
    assert event_a["event"] == "message", "Expected echo message event"
    payload_a = event_a["payload"]
    assert (payload_a.get("content") or payload_a.get("Content")) == "Hello Bob, this is Alice!", "Echo Content mismatch"

    print("\n--- 5. User B sends a Typing Indicator to User A ---")
    typing_payload = {
        "event": "typing",
        "payload": {
            "conversation_id": conv_id
        }
    }
    await ws_b.send(json.dumps(typing_payload))
    print("Bob sent typing indicator.")

    # Alice (User A) should receive the typing indicator
    msg_a_recv_typing = await ws_a.recv()
    event_a_typing = json.loads(msg_a_recv_typing)
    print(f"Alice received typing event: {event_a_typing}")
    assert event_a_typing["event"] == "typing", "Expected typing event"
    payload_a_typing = event_a_typing["payload"]
    assert (payload_a_typing.get("conversation_id") or payload_a_typing.get("ConversationID")) == conv_id, "Conversation ID mismatch"
    assert (payload_a_typing.get("user_id") or payload_a_typing.get("UserID")) == id_b, "User ID mismatch"

    print("\n--- 6. Cleanup & Disconnect ---")
    await ws_a.close()
    await ws_b.close()
    print("✓ All tests passed successfully!")

if __name__ == "__main__":
    asyncio.run(main())
