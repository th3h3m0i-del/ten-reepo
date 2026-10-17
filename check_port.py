import socket
import sys

def check_port(host, port):
    print(f"Đang thử mở cửa {host}:{port}...")
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(1)
    try:
        s.connect((host, port))
        print(f"[+] Cửa {port} đang mở.")
    except Exception:
        print(f"[-] Cửa {port} đóng hoặc bị chặn.")
    finally:
        s.close()

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Cách dùng: python check_port.py <host> <port>")
    else:
        check_port(sys.argv[1], int(sys.argv[2]))
