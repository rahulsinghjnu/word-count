# word-count
- This Application reads a large size file and breaks the large size of file into chunk of 64KB. 
- The 64KB chunk data is passed to word-count-service [https://github.com/rahulsinghjnu/word-count-service] to get the words frequency. 
- Chunk of data is being processed asynchronously and later results are being aggregated.
- The words are being sorted based on its frequency.
- Finally application returns top K words based on frequency in the given file as a json response.

# Prerequisites
- Make sure that word-count-service [https://github.com/rahulsinghjnu/word-count-service] rest service is up and running.
- Configure the word count rest service url into local.env file.

# How to run
- Execute the following command in order to get the TOP N words based on frequency in a large file.<br>
    `go run main.go -filePath=<FilePath>`
- TopN can be passed through command line argument. Default topN is 10.<br>
    `go run main.go -filePath=<FilePath> -topN=15`

# TODO
Error handling is todo in case of word-count-service is down.


