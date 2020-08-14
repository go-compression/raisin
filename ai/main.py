import os
import sys
import json


from helpers.generator import generate_pdf, generate_jpg, generate_jpg_params, generate_pdf_params
from helpers.ai import train

training_set = {
    (generate_pdf, 10, generate_pdf_params),
    (generate_jpg, 5, generate_jpg_params),
}

downloads = {
    ("http://corpus.canterbury.ac.nz/resources/large.zip", True),
    ("http://corpus.canterbury.ac.nz/resources/cantrbry.zip", True),
    ("http://corpus.canterbury.ac.nz/resources/artificl.zip", True),
    ("http://corpus.canterbury.ac.nz/resources/misc.zip", True),
    ("http://corpus.canterbury.ac.nz/resources/calgary.zip", True),
}

algorithms = ["arithmetic", "lzss", "flate", "gzip", "lzw", "zlib"]

benchmark_params = {
    "generate": False, 
    "download": False, 
    "fresh": False, 
    "delete_at_end": False,
}

def main(load_data=True, save_data=False, json_file="data.json"):
    if load_data:
        with open(json_file) as f:
            data = json.load(f)
    else:
        from helpers.compressor import benchmark
        data = benchmark(training_set, downloads, algorithms, **benchmark_params)

    if save_data:
        with open(json_file, 'w') as f:
            json.dump(data, f)
    
    train(data)

    print("Done")

if __name__ == "__main__":
    main()
