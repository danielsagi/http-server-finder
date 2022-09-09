import sys
import ipaddress

def main():
    with open("targets.txt", 'w') as f: 
        for ip in ipaddress.IPv4Network(sys.argv[1]):
            for protocol in ["http", "https"]:
                for port in [80, 443, 8080, 8000]:
                    f.write(f"{protocol}://{ip}:{port}\n")

if __name__ == '__main__':
    main()