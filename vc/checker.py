import argparse
import os
import re
import time

import subprocess


def run_command_with_limit(cmd, input_file, timeout):
    """
    Run a command with a time limit and redirecting stdin from a file.

    cmd: string, the command to run
    input_file: string, path to the file for stdin redirection
    timeout: int, time limit in seconds

    Returns: tuple containing return code, execution time, stdout, stderr,
             and a boolean indicating if the process was terminated due to a timeout
    """

    try:
        start_time = time.time()

        # If cmd is a string, split it into a list for subprocess.run
        if isinstance(cmd, str):
            cmd = cmd.split()

        with open(input_file, 'r') as f:
            completed_process = subprocess.run(cmd, stdin=f, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                               timeout=timeout)

        stdout = completed_process.stdout.decode('utf-8')
        stderr = completed_process.stderr.decode('utf-8')

        return_code = completed_process.returncode
        end_time = time.time()
        was_timeout = False

    except subprocess.TimeoutExpired:
        # This block will be entered if the command times out
        end_time = time.time()
        stdout = ''
        stderr = ''
        return_code = ''
        was_timeout = True

    except subprocess.CalledProcessError as e:
        # This block will be entered if the command fails
        end_time = time.time()
        stdout = e.stdout.decode('utf-8') if e.stdout else ''
        stderr = e.stderr.decode('utf-8') if e.stderr else ''
        return_code = e.returncode
        was_timeout = False

    except Exception as e:
        # This block will be entered if an unexpected exception occurs
        end_time = time.time()
        stdout = ''
        stderr = str(e)
        return_code = -1
        was_timeout = False

    return return_code, end_time - start_time, stdout, stderr, was_timeout


def extract_starting_numerical_prefix(filename):
    number = ''.join(filter(str.isdigit, filename))
    return int(number) if number else 0


def get_grouped_files(in_dir):
    groups = {}

    sorted_files = sorted(os.listdir(in_dir))
    sorted_files = [f for f in sorted_files if f.endswith(".in")]

    flatten = True

    for f in sorted_files:
        group = extract_starting_numerical_prefix(f)
        if group not in groups:
            groups[group] = [f]
        else:
            flatten = False
            groups[group].append(f)

    if flatten:
        return {'0': sorted_files}
    return groups


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("executable", type=str, help="Path to the executable")
    parser.add_argument("--time_limit", type=int, default=60, help="Time limit [sec] (default: 60)")
    parser.add_argument("--max_time_limit_exceeded", type=int, default=10, help="Max time limit exceeded (default: 10)")

    args = parser.parse_args()

    executable = args.executable
    time_limit = args.time_limit
    max_time_limit_exceeded = args.max_time_limit_exceeded

    in_dir = "in"
    out_dir = "out"
    checker_file = "chk.py"
    print("file,status,time,return,stderr")

    grouped_files = get_grouped_files(in_dir)
    found_error = False
    for group in grouped_files.values():
        tles = 0
        for f in group:
            in_file = os.path.join(in_dir, f)
            return_code, time, stdout, stderr, was_timeout = run_command_with_limit(executable, in_file, time_limit)
            time = "{:.3f}".format(time)

            if not was_timeout and return_code == 0:
                out_file = os.path.join(out_dir, f.replace(".in", ".out"))

                if os.path.exists(checker_file):
                    with open(".user_out.txt", 'w') as result:
                        result.write(stdout)
                    result = subprocess.run(["python3", checker_file, in_file, ".user_out.txt", out_file],
                                            stdout=subprocess.PIPE, stderr=subprocess.PIPE)
                    stdout = result.stdout.decode('utf-8')

                    if "OK" in stdout:
                        status = "OK"
                    else:
                        status = "Wrong"
                        stderr += "\n" + stdout + "\n" + stderr
                else:
                    with open(out_file, 'r') as result:
                        solution = result.read()
                        if solution.strip() == stdout.strip():
                            status = "OK"
                        else:
                            status = "Wrong"
            elif was_timeout:
                status = "timelimit"
                time = ''
            else:
                stderr += "\nNon zero exit code"
                status = "Wrong"

            stderr = re.sub(r"\n", r"\\n", stderr)
            print(f"{f},{status},{time},{return_code},{stderr}", flush=True)

            if status == "Wrong":
                found_error = True
                break
            if status == "timelimit":
                tles += 1
                if tles >= max_time_limit_exceeded:
                    break
        if found_error:
            break

    try:
        os.remove(".user_out.txt")
    except FileNotFoundError:
        pass


if __name__ == '__main__':
    main()
