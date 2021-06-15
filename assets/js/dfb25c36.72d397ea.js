(self.webpackChunkbuildbuddy_docs_website=self.webpackChunkbuildbuddy_docs_website||[]).push([[2187],{3905:function(e,t,n){"use strict";n.d(t,{Zo:function(){return p},kt:function(){return f}});var r=n(7294);function o(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function i(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function a(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?i(Object(n),!0).forEach((function(t){o(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):i(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function l(e,t){if(null==e)return{};var n,r,o=function(e,t){if(null==e)return{};var n,r,o={},i=Object.keys(e);for(r=0;r<i.length;r++)n=i[r],t.indexOf(n)>=0||(o[n]=e[n]);return o}(e,t);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(r=0;r<i.length;r++)n=i[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(o[n]=e[n])}return o}var s=r.createContext({}),c=function(e){var t=r.useContext(s),n=t;return e&&(n="function"==typeof e?e(t):a(a({},t),e)),n},p=function(e){var t=c(e.components);return r.createElement(s.Provider,{value:t},e.children)},d={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},u=r.forwardRef((function(e,t){var n=e.components,o=e.mdxType,i=e.originalType,s=e.parentName,p=l(e,["components","mdxType","originalType","parentName"]),u=c(n),f=o,m=u["".concat(s,".").concat(f)]||u[f]||d[f]||i;return n?r.createElement(m,a(a({ref:t},p),{},{components:n})):r.createElement(m,a({ref:t},p))}));function f(e,t){var n=arguments,o=t&&t.mdxType;if("string"==typeof e||o){var i=n.length,a=new Array(i);a[0]=u;var l={};for(var s in t)hasOwnProperty.call(t,s)&&(l[s]=t[s]);l.originalType=e,l.mdxType="string"==typeof e?e:o,a[1]=l;for(var c=2;c<i;c++)a[c]=n[c];return r.createElement.apply(null,a)}return r.createElement.apply(null,n)}u.displayName="MDXCreateElement"},230:function(e,t,n){"use strict";n.r(t),n.d(t,{frontMatter:function(){return a},metadata:function(){return l},toc:function(){return s},default:function(){return p}});var r=n(2122),o=n(9756),i=(n(7294),n(3905)),a={id:"config-storage",title:"Storage Configuration",sidebar_label:"Storage"},l={unversionedId:"config-storage",id:"config-storage",isDocsHomePage:!1,title:"Storage Configuration",description:"storage: The Storage section configures where and how BuildBuddy will store blob data. Required",source:"@site/../docs/config-storage.md",sourceDirName:".",slug:"/config-storage",permalink:"/docs/config-storage",editUrl:"https://github.com/buildbuddy-io/buildbuddy/edit/master/docs/../docs/config-storage.md",version:"current",sidebar_label:"Storage",frontMatter:{id:"config-storage",title:"Storage Configuration",sidebar_label:"Storage"},sidebar:"someSidebar",previous:{title:"Database Configuration",permalink:"/docs/config-database"},next:{title:"Cache Configuration",permalink:"/docs/config-cache"}},s=[{value:"Options",id:"options",children:[]},{value:"Example sections",id:"example-sections",children:[{value:"Disk",id:"disk",children:[]},{value:"GCS",id:"gcs",children:[]},{value:"AWS S3",id:"aws-s3",children:[]}]}],c={toc:s};function p(e){var t=e.components,n=(0,o.Z)(e,["components"]);return(0,i.kt)("wrapper",(0,r.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,i.kt)("p",null,(0,i.kt)("inlineCode",{parentName:"p"},"storage:")," The Storage section configures where and how BuildBuddy will store blob data. ",(0,i.kt)("strong",{parentName:"p"},"Required")),(0,i.kt)("h2",{id:"options"},"Options"),(0,i.kt)("p",null,"One of the following sections is ",(0,i.kt)("strong",{parentName:"p"},"Required")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},(0,i.kt)("inlineCode",{parentName:"p"},"disk:")," The Disk section configures disk-based blob storage."),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("inlineCode",{parentName:"li"},"root_directory")," The root directory to store all blobs in, if using disk based storage. This directory must be readable and writable by the BuildBuddy process. The directory will be created if it does not exist."))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},(0,i.kt)("inlineCode",{parentName:"p"},"gcs:")," The GCS section configures Google Cloud Storage based blob storage."),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},(0,i.kt)("inlineCode",{parentName:"p"},"bucket")," The name of the GCS bucket to store files in. Will be created if it does not already exist.")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},(0,i.kt)("inlineCode",{parentName:"p"},"credentials_file")," A path to a ",(0,i.kt)("a",{parentName:"p",href:"https://cloud.google.com/docs/authentication/getting-started"},"JSON credentials file")," that will be used to authenticate to GCS.")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},(0,i.kt)("inlineCode",{parentName:"p"},"project_id")," The Google Cloud project ID of the project owning the above credentials and GCS bucket.")))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},(0,i.kt)("inlineCode",{parentName:"p"},"aws_s3:")," The AWS section configures AWS S3 storage."),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},(0,i.kt)("inlineCode",{parentName:"p"},"region")," The AWS region")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},(0,i.kt)("inlineCode",{parentName:"p"},"bucket")," The AWS S3 bucket (will be created automatically)")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},(0,i.kt)("inlineCode",{parentName:"p"},"credentials_profile")," If a profile other than default is chosen, use that one.")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"By default, the S3 blobstore will rely on environment variables, shared credentials, or IAM roles. See ",(0,i.kt)("a",{parentName:"p",href:"https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials"},"AWS Go SDK docs")," for more information."))))),(0,i.kt)("p",null,(0,i.kt)("strong",{parentName:"p"},"Optional")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("inlineCode",{parentName:"li"},"chunk_file_size_bytes:")," How many bytes to buffer in memory before flushing a chunk of build protocol data to disk.")),(0,i.kt)("h2",{id:"example-sections"},"Example sections"),(0,i.kt)("h3",{id:"disk"},"Disk"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},"storage:\n  ttl_seconds: 86400  # One day in seconds.\n  chunk_file_size_bytes: 3000000  # 3 MB\n  disk:\n    root_directory: /tmp/buildbuddy\n")),(0,i.kt)("h3",{id:"gcs"},"GCS"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},'storage:\n  ttl_seconds: 0  # No TTL.\n  chunk_file_size_bytes: 3000000  # 3 MB\n  gcs:\n    bucket: "buildbuddy_blobs"\n    project_id: "my-cool-project"\n    credentials_file: "enterprise/config/my-cool-project-7a9d15f66e69.json"\n')),(0,i.kt)("h3",{id:"aws-s3"},"AWS S3"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},'storage:\n  aws_s3:\n    # required\n    region: "us-west-2"\n    bucket: "buddybuild-bucket"\n    # optional\n    credentials_profile: "other-profile"\n')))}p.isMDXComponent=!0}}]);