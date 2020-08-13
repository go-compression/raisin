import os
import sys
import math
import uuid

import magic
from pathlib import Path

from generator import generate_pdf, generate_image

# Add the engine directory to the path of importable directories
engine_path = os.getcwd() + "/engine"
sys.path.append(engine_path)
from engine import engine

class FileData():
    def __init__(self, filename, entropy=None, filetype=None) -> None:
        self.filename = filename
        if not entropy:
            # Credit for entropy calculations: https://github.com/mattnotmax/entropy/blob/master/entropy.py
            with open(filename, 'rb') as f:
                byteArr = list(f.read())
            fileSize = len(byteArr)
            freqList = []
            for b in range(256):
                ctr = 0
                for byte in byteArr:
                    if byte == b:
                        ctr += 1
                freqList.append(float(ctr) / fileSize)
            # Shannon entropy
            ent = 0.0
            for freq in freqList:
                if freq > 0:
                    ent = ent + freq * math.log(freq, 2)
            entropy = -ent

        self.entropy = entropy
        
        if not filetype:
            f = magic.Magic(mime=True, uncompress=True)
            filetype = f.from_file(filename)

        self.filetype = filetype
    
    def delete(self):
        os.remove(self.filename)


class CompressionResult():
    def __init__(self, file_data, result) -> None:
        self.file_data = file_data
        self.result = result

training_set = {
    generate_pdf: 100,
    generate_image: 10,
}

def main():
    files = []
    Path("files").mkdir(parents=True, exist_ok=True)
    os.chdir("files")



    for generator, times in training_set.items():
        for i in range(times):
            name = str(uuid.uuid4())
            text = name

            file = generator(name, text)

            data = FileData(file)
            files.append(data)

    try:
        benchmark_files(files)
    except Exception as e:
        print("Exception occured, cleaning up files")
        for file in files:
            file.delete()
        raise e
    
    for file in files:
        file.delete()
    
    print("Done!")

def benchmark_files(files):
    data = []
    settings = engine.Settings(WriteOutFiles=False, PrintStatus=False, PrintStats=False)

    for file in files:
        result = engine.BenchmarkFile("arithmetic", file.filename, settings)
        data.append(CompressionResult(file, result))

if __name__ == "__main__":
    main()
