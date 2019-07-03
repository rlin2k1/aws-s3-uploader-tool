/* main.go
Web Application to Provide an easy and simple way for Poshmark Marketing Team to
upload images to Amazon S3 for marketing campaigns.

Author(s):
	Roy Lin

Date Created:
	June 25th, 2019
*/

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	mapset "github.com/deckarep/golang-set"
)

//var templates = template.Must(template.ParseFiles("index.html"))         //MUST Panics if ParseFiles returns ERROR
var validPath = regexp.MustCompile("/.(gif|jpg|jpeg|tiff|png|txt|doc|docx)$/") //CHECK IF IMAGE

type Page struct {
	CurrentFolder  string
	FolderContents []string
	FileContents   []string
}

// UploadFileToS3 saves a file to aws bucket and returns the url to // the file and an error if there's any
func UploadFileToS3(s *session.Session, file multipart.File, fileHeader *multipart.FileHeader, folderName string) (string, error) {
	// get the file size and read
	// the file content into a buffer
	size := fileHeader.Size
	buffer := make([]byte, size)
	file.Read(buffer)
	if folderName == "Root" {
		folderName = ""
	}
	// create a unique file name for the file
	tempFileName := folderName + "/" + fileHeader.Filename

	// config settings: this is where you choose the bucket,
	// filename, content-type and storage class of the file
	// you're uploading
	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String("bucket-upload-roy1"),
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String("public-read"), // could be private if you want it to be access by only authorized users
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(int64(size)),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})
	fmt.Println("FILENAME: ", aws.String(tempFileName))
	if err != nil {
		return "", err
	}

	return tempFileName, err
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		http.Redirect(w, r, "/", http.StatusFound)
		fmt.Println("CAUGHT")
	} else if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			fmt.Println("CAUGHT")
			return
		}
		defer file.Close()

		// create an AWS session which can be
		// reused if we're uploading many files
		s, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-2"),
			Credentials: credentials.NewStaticCredentials(
				"AKIA2EJR7MOTGUC644N2",                     // id
				"p0M/OWDdydpwg41yXPm3ztZSRYSXKRM++C8rxz1h", // secret
				""), // token can be left blank for now
		})
		svc := s3.New(s)

		bucket := "bucket-upload-roy1"

		resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
		if err != nil {
			return
		}

		set := mapset.NewSet()

		for _, item := range resp.Contents {
			fmt.Println(*item.Key)
			if strings.Contains(*item.Key, "/") {
				directory := strings.Split(*item.Key, "/")[0]
				if !set.Contains(directory) {
					fmt.Println("Directory Name:         ", directory)
					fmt.Println("")
					set.Add(directory)
				}
			}
		}

		if err != nil {
			fmt.Fprintf(w, "Could not upload file")
		}

		fileName, err := UploadFileToS3(s, file, handler, r.FormValue("token"))
		if err != nil {
			fmt.Fprintf(w, "Could not upload file")
		}
		fmt.Println("Uploaded to Folder:", fileName)
		fmt.Fprintf(w, "<div>Image Uploaded Successfully: %v</div><h1>[<a href=\"/\">Upload Another File</a>]</h1>", handler.Filename)
		//UPLOAD ANOTHER IMAGE FROM WIKI
	} else {
		http.Error(w, "GET OR POST", http.StatusInternalServerError)
		return
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[len("/"):]
	s := []string{}
	f := []string{}
	s1, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(
			"AKIA2EJR7MOTGUC644N2",                     // id
			"p0M/OWDdydpwg41yXPm3ztZSRYSXKRM++C8rxz1h", // secret
			""), // token can be left blank for now
	})
	svc := s3.New(s1)

	bucket := "bucket-upload-roy1"

	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
	if err != nil {
		return
	}

	set := mapset.NewSet()

	for _, item := range resp.Contents {
		fmt.Println(*item.Key, *item.Size, filePath)
		if *item.Size == 0 && strings.Contains(*item.Key, filePath) {
			newPath := (*item.Key)[len(filePath):]
			//fmt.Println("Directory Name1:         ", newPath[:strings.IndexByte(newPath, '/')])
			newPath = strings.Split(newPath, "/")[0] + "/"
			if !set.Contains(newPath) && newPath != "/" {
				fmt.Println("Directory Name:         ", newPath)
				fmt.Println("")
				set.Add(newPath)
				s = append(s, newPath[:len(newPath)-1])
			}
		} else if strings.Contains(*item.Key, filePath) {
			newPath := (*item.Key)[len(filePath):]
			if !strings.Contains(newPath, "/") {
				f = append(f, newPath)
			}
		}
	}
	if filePath == "" {
		filePath = "/"
	}
	t, _ := template.ParseFiles("src/index.html")
	page := &Page{CurrentFolder: filePath, FolderContents: s, FileContents: f}
	t.Execute(w, page)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	fmt.Println("method:", r.Method)
	s, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(
			"AKIA2EJR7MOTGUC644N2",                     // id
			"p0M/OWDdydpwg41yXPm3ztZSRYSXKRM++C8rxz1h", // secret
			""), // token can be left blank for now
	})
	svc := s3.New(s)

	bucket := "bucket-upload-roy1"
	list := r.Form["int"]
	fmt.Println(list)
	// create a unique file name for the file

	for _, item := range list {
		svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(item)})

		svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	}
	// config settings: this is where you choose the bucket,
	// filename, content-type and storage class of the file
	// you're uploading
	fmt.Fprintf(w, "<div>Deletion Successful!</div><h1>[<a href=\"/\">Upload A File</a>]</h1>")
}

func main() {
	fmt.Println("Starting Web Server")

	http.Handle("/image_assets/", http.StripPrefix("/image_assets/",
		http.FileServer(http.Dir("image_assets"))))
	http.Handle("/src/", http.StripPrefix("/src/",
		http.FileServer(http.Dir("src"))))

	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/upload/", uploadHandler)
	http.HandleFunc("/", rootHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
