import sys
import ipaddress

def main():
    with open("targets.txt", 'w') as f: 
        for ip in ipaddress.IPv4Network(sys.argv[1]):
            for protocol, port in [("http", 80), ("https", 443)]:
                    f.write(f"{protocol}://{ip}:{port}\n")

if __name__ == '__main__':
    main()