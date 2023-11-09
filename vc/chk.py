import os.path
import sys


def read_solution_size():
    with open(sys.argv[3], 'r') as f:
        try:
            sol_size = int(f.readline().strip())  # Read the first line, strip whitespace and convert to integer
            return sol_size  # Return the size of the optimal solution
        except ValueError:  # If the conversion to integer fails
            print("Error: Can not read solution size from model output file")
            sys.exit(1)  # Exit the program


class Graph:
    def __init__(self):
        self.edges = set()
        self.usedEdges = set()


def main():
    if len(sys.argv) != 4:
        print("Usage: python vcchk.py <input_file> <user_output_file> <model_output_file>")
        sys.exit(1)

    if not os.path.exists(sys.argv[2]):
        print("WRONG\nUser output file does not exist")
        sys.exit(1)

    sol_size = read_solution_size()

    vc = set()
    with open(sys.argv[2], 'r') as f:
        for line in f.readlines():
            line = line.split("#")[0].strip()
            if line == '':
                continue  # Skip empty lines

            vc.add(line)

    with open(sys.argv[1], 'r') as f:
        f.readline()
        for line in f.readlines():
            edge = tuple(line.split())
            if edge[0] not in vc and edge[1] not in vc:
                print("WRONG\nEdge not covered")
                sys.exit(1)

    if len(vc) > sol_size:
        print("WRONG\nToo many nodes")
        sys.exit(1)

    print("OK")


if __name__ == '__main__':
    main()
