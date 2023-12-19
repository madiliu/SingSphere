# SingSphere

## Overview

[SingSphere](http://singsphere.s3-website-us-east-1.amazonaws.com) is an innovative online karaoke platform that integrates advanced cloud services for a seamless and interactive singing experience. It focuses on social interaction, allowing users to sing synchronously with friends across the globe.

## Key Features

- Real-time audio synchronization in a distributed environment.
- Scalable and responsive solution using AWS Lambda, EC2, OpenSearch, and RabbitMQ.
- Focuses on the social aspects of karaoke, connecting friends who are physically apart.
<img width="1512" alt="Screenshot 2023-12-18 at 10 39 00 PM" src="https://github.com/madiliu/SingSphere/assets/90917906/5eed39ba-30f8-4352-b407-72cc1b6818be">
<img width="1512" alt="Screenshot 2023-12-18 at 10 39 08 PM" src="https://github.com/madiliu/SingSphere/assets/90917906/a3a9363b-6c27-429b-afca-37df5ee20ea7">
<img width="1512" alt="Screenshot 2023-12-18 at 10 39 18 PM" src="https://github.com/madiliu/SingSphere/assets/90917906/a336e232-4ed1-4aca-ae98-5b44620d95f7">
<img width="1512" alt="Screenshot 2023-12-18 at 10 39 47 PM" src="https://github.com/madiliu/SingSphere/assets/90917906/ea3efa91-db4b-4d00-b87a-5756bc955e73">

## Design Architecture
<img width="1496" alt="Screenshot 2023-12-18 at 6 52 54 PM" src="https://github.com/Nahemah1022/SingSphere/assets/90917906/44ed765d-aab4-47fc-9914-f1cc241c544b">

- Utilizes AWS’s content delivery network service, CloudFront, and S3 Frontend for low latency.
- API Gateway for routing requests and providing HTTPS endpoints.
- Lambda Functions for indexing and searching music, and enqueueing song requests.
- RabbitMQ for managing communication and queuing song requests.
- OpenSearch for efficient song metadata search and retrieval.
- Codepipeline for automation release of codes.

## Technology Used

- Frontend: React.js for dynamic user interfaces.
- Backend: Go for handling concurrent operations and Python for AWS Lambda functions.
- AWS services for scalable cloud infrastructure.

## Key Findings and Future Work

- Performance tests highlighted challenges in real-time audio synchronization.
- Future work focuses on scaling the system, implementing advanced load balancing, and exploring innovative synchronization techniques.

## Conclusion

SingSphere demonstrates the potential for cloud-based solutions in real-time entertainment applications, emphasizing the importance of scalability and real-time interaction in digital platforms.

## Contributors
[Yin-Tung Chen](https://github.com/Nahemah1022), [Josephine Chan](https://github.com/honey-grapes), [Madelain Liu](https://github.com/madiliu), Hsin-Yuang Huang
