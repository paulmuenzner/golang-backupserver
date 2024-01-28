package services

import (
	config "backupserver/config"
	aws "backupserver/utils/aws"
	csvHandler "backupserver/utils/csv"
	"fmt"
)

func UploadBackupsAwsS3(folderPathBackup, metaFileName string, awsClientConfig *aws.AwsClientConfigData, bucketName string) error {

	// Create an instance of RowTypesMeta
	rowMetaInstance := config.RowTypesMeta{}

	// Open related meta file retrieving information of all backup files
	filePathMeta := folderPathBackup + "/" + metaFileName
	metaDataAsMap, err := csvHandler.ConvertCsvToMap(filePathMeta, rowMetaInstance)
	if err != nil {
		return fmt.Errorf("Couldn't convert csv to map for file path %s in 'UploadBackupsAwsS3' with 'ConvertCsvToMap()'. Error: %v", filePathMeta, err)
	}

	// Setup AWS S3 client dependency
	awsMethods, err := aws.GetAwsMethods(awsClientConfig)
	if err != nil {
		return fmt.Errorf("Error in 'UploadBackupsAwsS3()' with 'AwsMethodInterface()'. Error:  %v", bucketName, err)
	}

	// Validate if bucket accessible
	bucketExist, err := awsMethods.MethodInterface.BucketExists(bucketName)
	if !bucketExist {
		return fmt.Errorf("S3 Bucket %s not found in 'UploadBackupsAwsS3' with 'BucketExists(bucketName)'. Error:  %v", bucketName, err)
	}
	if err != nil {
		return fmt.Errorf("Error in 'UploadBackupsAwsS3' validating if bucket '%s' accessible with 'BucketExists(bucketName)'. Error:  %v", bucketName, err)
	}

	// Loop through each backup file row listed in meta file
	for _, value := range metaDataAsMap {
		filePath := value["folder_path"] + "/" + value["file_name"] + ".csv"

		err = awsMethods.MethodInterface.UploadFile(bucketName, filePath, filePath)

		if err != nil {
			return fmt.Errorf("Couldn't upload file in 'UploadBackupsAwsS3' with 'UploadFile()'. Error: %v", err)
		}
	}

	// Add meta data file itself to backup path on S3
	err = awsMethods.MethodInterface.UploadFile(bucketName, filePathMeta, filePathMeta)
	if err != nil {
		return fmt.Errorf("Couldn't upload meta file of path '%s' in 'UploadBackupsAwsS3' with 'UploadFile()'. Error: %v", filePathMeta, err)
	}

	// Circular buffer S3 - Only store latest number of n backups and delete older ones if circular buffer activated
	isCircularBufferActivatedS3 := config.IsCircularBufferActivatedS3
	if isCircularBufferActivatedS3 == true {
		err := DeleteOldBackupsS3(bucketName, awsClientConfig)
		if err != nil {
			return fmt.Errorf("Error in 'Backup' applying 'DeleteOldBackups()'. Error: %v", err)

		}
	}

	return nil
}
