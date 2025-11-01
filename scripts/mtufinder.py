import ping3
import time
import sys

def test_mtu(host, mtu_size):
    """
    Test a specific MTU size by sending a ping with the corresponding payload.
    :param host: Destination IP address
    :param mtu_size: MTU size to test
    :return: True if ping is successful, False otherwise
    """
    try:
        # Calculate payload size (MTU minus 28 bytes for IP and ICMP headers)
        payload_size = mtu_size - 28
        if payload_size < 0:
            return False

        # Send ping with specified payload size
        response_time = ping3.ping(host, size=payload_size, timeout=2)
        
        # Check if response is valid
        if response_time is not None and response_time >= 0:
            return True
        return False
    except Exception as e:
        print(f"Error testing MTU {mtu_size}: {e}")
        return False

def find_best_mtu(host, min_mtu, max_mtu, step):
    """
    Find the best MTU by testing a range of sizes.
    :param host: Destination IP address
    :param min_mtu: Minimum MTU to test
    :param max_mtu: Maximum MTU to test
    :param step: Step size for MTU increments
    :return: Best MTU found
    """
    print(f"Testing MTU for {host}...")
    best_mtu = min_mtu
    for mtu in range(min_mtu, max_mtu + 1, step):
        print(f"Testing MTU: {mtu}")
        if test_mtu(host, mtu):
            best_mtu = mtu
            print(f"MTU {mtu} successful")
        else:
            print(f"MTU {mtu} failed")
            break
        time.sleep(1)  # Delay between tests to avoid network overload
    return best_mtu

def validate_ip(host):
    """
    Validate if the input is a valid IP address or hostname.
    :param host: Input string to validate
    :return: True if valid, False otherwise
    """
    try:
        # Try resolving hostname or IP
        import socket
        socket.gethostbyname(host)
        return True
    except socket.error:
        return False

def get_user_input():
    """
    Collect user inputs interactively for the MTU test.
    :return: Tuple of (host, min_mtu, max_mtu, step)
    """
    print("=== MTU Testing Wizard ===")
    
    # Get destination host
    while True:
        host = input("Enter the destination host (e.g., 8.8.8.8 or your VPN server IP): ").strip()
        if validate_ip(host):
            break
        print("Invalid IP address or hostname. Please try again.")

    # Get minimum MTU
    while True:
        try:
            min_mtu = int(input("Enter the minimum MTU to test (default 576): ") or 576)
            if min_mtu >= 68:  # Minimum MTU per IPv4 standard
                break
            print("Minimum MTU must be at least 68.")
        except ValueError:
            print("Please enter a valid number.")

    # Get maximum MTU
    while True:
        try:
            max_mtu = int(input("Enter the maximum MTU to test (default 1500): ") or 1500)
            if max_mtu >= min_mtu:
                break
            print("Maximum MTU must be greater than or equal to minimum MTU.")
        except ValueError:
            print("Please enter a valid number.")

    # Get step size
    while True:
        try:
            step = int(input("Enter the step size for MTU increments (default 10): ") or 10)
            if step > 0:
                break
            print("Step size must be a positive number.")
        except ValueError:
            print("Please enter a valid number.")

    return host, min_mtu, max_mtu, step

def main():
    try:
        # Collect user inputs
        host, min_mtu, max_mtu, step = get_user_input()

        # Run MTU test
        best_mtu = find_best_mtu(host, min_mtu, max_mtu, step)
        
        # Display results
        print(f"\nBest MTU found: {best_mtu}")
        print("Note: Apply this MTU value in your VPN or network interface settings and verify performance.")
        
    except KeyboardInterrupt:
        print("\nTest interrupted by user.")
        sys.exit(1)
    except Exception as e:
        print(f"An error occurred: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()