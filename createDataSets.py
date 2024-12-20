import random
import os

# Directory to save the files
directory = "random_numbers_files"

# Create the directory if it doesn't exist
os.makedirs(directory, exist_ok=True)

# Number of files to generate
num_files = 1000
# Numbers per file
numbers_per_file = 100000

for file_index in range(1, num_files + 1):
    # File path
    file_path = os.path.join(directory, f"{file_index}.txt")
    
    with open(file_path, "w") as file:
        for _ in range(numbers_per_file):
            # Generate a random 4-digit number
            number = random.randint(1000, 9999)
            file.write(f"{number}\n")

print(f"{num_files} files have been created in the '{directory}' directory.")
