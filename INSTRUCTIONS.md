# Beam Transfer Coding Challenge

## Introduction

This assignment simulates a real-world scenario where files of varying sizes must be processed, packaged, and transferred securely to an object storage endpoint. 
Your task is to design an algorithm and implement a Go project that organizes small files into manageable and efficient 60MB blocks while ensuring data integrity and preserving their structure upon retrieval. 

## Sample Files

Use the provided script to generate file structure. Refer to README for instructions on how to use the file generator.

[test-files.zip](https://prod-files-secure.s3.us-west-2.amazonaws.com/9b3fda7f-a489-411d-99a5-668e41d27764/cef3678e-1cc7-4aa1-a67e-e1f814c0659d/test-files.zip)

```go
go run main.go sample-files.yml
```

## Exercise

Your task is mainly to design algorithm and implement it together with simple project in Go that performs the following:

1. **Input Sample Files**:
    - A folder containing multiple files of varying sizes (e.g., text files, binary files, etc.).
        - Refer to Sample Files, on how to generate test files set.
    - The files can range from a few KBs to several MBs.
2. **Objective**:
    - Efficiently pack files into **blocks of 60MB.**
        - Files larger than 60MB do not need to be processed.
        - Consider any kind of archive which is most suitable for task and time scope of assignment.
    - Propose basic algorithm for such a task.
        - Simple algorithm is good enough for this purpose. But proposals for additional improvements is mandatory.
    - Make sure you do put enough effort to methods which are related to data reading, an processing, and consider proper optimisation of them.
    - Take in account the data integrity for possible data transfer an retrieval.
3. **Output**:
    - **We do not need to implement actual upload/download methods for any type of object storage.**
    - Ensure that you do design proper methods that verify uploaded data upon download and can be:
        - **Unpacked** into the original file structure
        - Verified for integrity. (optional)

## Delivery

Your result should be delivered in a public Github repository that also exposes the development process. Use any kind of best development practices for project you would like to demonstrate as suitable.
