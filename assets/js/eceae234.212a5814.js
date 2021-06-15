(self.webpackChunkbuildbuddy_docs_website=self.webpackChunkbuildbuddy_docs_website||[]).push([[8484],{3905:function(e,t,r){"use strict";r.d(t,{Zo:function(){return s},kt:function(){return m}});var n=r(7294);function o(e,t,r){return t in e?Object.defineProperty(e,t,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[t]=r,e}function a(e,t){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),r.push.apply(r,n)}return r}function l(e){for(var t=1;t<arguments.length;t++){var r=null!=arguments[t]?arguments[t]:{};t%2?a(Object(r),!0).forEach((function(t){o(e,t,r[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):a(Object(r)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(r,t))}))}return e}function i(e,t){if(null==e)return{};var r,n,o=function(e,t){if(null==e)return{};var r,n,o={},a=Object.keys(e);for(n=0;n<a.length;n++)r=a[n],t.indexOf(r)>=0||(o[r]=e[r]);return o}(e,t);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(n=0;n<a.length;n++)r=a[n],t.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(o[r]=e[r])}return o}var u=n.createContext({}),d=function(e){var t=n.useContext(u),r=t;return e&&(r="function"==typeof e?e(t):l(l({},t),e)),r},s=function(e){var t=d(e.components);return n.createElement(u.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return n.createElement(n.Fragment,{},t)}},c=n.forwardRef((function(e,t){var r=e.components,o=e.mdxType,a=e.originalType,u=e.parentName,s=i(e,["components","mdxType","originalType","parentName"]),c=d(r),m=o,h=c["".concat(u,".").concat(m)]||c[m]||p[m]||a;return r?n.createElement(h,l(l({ref:t},s),{},{components:r})):n.createElement(h,l({ref:t},s))}));function m(e,t){var r=arguments,o=t&&t.mdxType;if("string"==typeof e||o){var a=r.length,l=new Array(a);l[0]=c;var i={};for(var u in t)hasOwnProperty.call(t,u)&&(i[u]=t[u]);i.originalType=e,i.mdxType="string"==typeof e?e:o,l[1]=i;for(var d=2;d<a;d++)l[d]=r[d];return n.createElement.apply(null,l)}return n.createElement.apply(null,r)}c.displayName="MDXCreateElement"},253:function(e,t,r){"use strict";r.r(t),r.d(t,{frontMatter:function(){return l},metadata:function(){return i},toc:function(){return u},default:function(){return s}});var n=r(2122),o=r(9756),a=(r(7294),r(3905)),l={id:"on-prem",title:"On-prem Quickstart",sidebar_label:"On-prem Quickstart"},i={unversionedId:"on-prem",id:"on-prem",isDocsHomePage:!1,title:"On-prem Quickstart",description:"BuildBuddy is designed to be easy to run on-premise for those use cases where data absolutely must not leave a company's servers. It can be run your own servers, or in your own cloud environment. It supports major cloud providers like GCP, AWS, and Azure.",source:"@site/../docs/on-prem.md",sourceDirName:".",slug:"/on-prem",permalink:"/docs/on-prem",editUrl:"https://github.com/buildbuddy-io/buildbuddy/edit/master/docs/../docs/on-prem.md",version:"current",sidebar_label:"On-prem Quickstart",frontMatter:{id:"on-prem",title:"On-prem Quickstart",sidebar_label:"On-prem Quickstart"},sidebar:"someSidebar",previous:{title:"Cloud Quickstart",permalink:"/docs/cloud"},next:{title:"Contributing to BuildBuddy",permalink:"/docs/contributing"}},u=[{value:"Getting started",id:"getting-started",children:[]},{value:"Bazel Run",id:"bazel-run",children:[]},{value:"Docker Image",id:"docker-image",children:[]},{value:"Kubernetes",id:"kubernetes",children:[{value:"Custom configuration",id:"custom-configuration",children:[]},{value:"Output to yaml file",id:"output-to-yaml-file",children:[]},{value:"Number of replicas",id:"number-of-replicas",children:[]},{value:"Restart behavior",id:"restart-behavior",children:[]},{value:"Enterprise deployment",id:"enterprise-deployment",children:[]}]},{value:"Helm",id:"helm",children:[]},{value:"Configuring BuildBuddy",id:"configuring-buildbuddy",children:[]}],d={toc:u};function s(e){var t=e.components,r=(0,o.Z)(e,["components"]);return(0,a.kt)("wrapper",(0,n.Z)({},d,r,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("p",null,"BuildBuddy is designed to be easy to run on-premise for those use cases where data absolutely must not leave a company's servers. It can be run your own servers, or in your own cloud environment. It supports major cloud providers like GCP, AWS, and Azure."),(0,a.kt)("p",null,"The software itself is open-source and easy to audit."),(0,a.kt)("p",null,"For companies, we offer an ",(0,a.kt)("a",{parentName:"p",href:"/docs/enterprise"},"Enterprise")," version of BuildBuddy that contains advanced features like OIDC Auth, API access, and more."),(0,a.kt)("h2",{id:"getting-started"},"Getting started"),(0,a.kt)("p",null,"There are four ways to run BuildBuddy on-prem:"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"#bazel-run"},"Bazel Run"),": get the source and run a simple ",(0,a.kt)("inlineCode",{parentName:"li"},"bazel run")," command."),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"#docker-image"},"Docker Image"),": pre-built Docker images running the latest version of BuildBuddy."),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"#kubernetes"},"Kubernetes"),": deploy BuildBuddy to your Kubernetes cluster with a one-line deploy script."),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"#helm"},"Helm"),": deploy BuildBuddy to your Kubernetes cluster with the official BuildBuddy helm charts.")),(0,a.kt)("h2",{id:"bazel-run"},"Bazel Run"),(0,a.kt)("p",null,'The simplest method of running BuildBuddy on your own computer is to download and run it with "bazel run". Doing that is simple:'),(0,a.kt)("ol",null,(0,a.kt)("li",{parentName:"ol"},"Get the source")),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},'git clone "https://github.com/buildbuddy-io/buildbuddy"\n')),(0,a.kt)("ol",{start:2},(0,a.kt)("li",{parentName:"ol"},"Navigate into the BuildBuddy directory")),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},"cd buildbuddy\n")),(0,a.kt)("ol",{start:3},(0,a.kt)("li",{parentName:"ol"},"Build and run using bazel")),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},"bazel run -c opt server:buildbuddy\n")),(0,a.kt)("p",null,"We recommend using a tool like ",(0,a.kt)("a",{parentName:"p",href:"https://github.com/bazelbuild/bazelisk"},"Bazelisk")," that respects the repo's ",(0,a.kt)("a",{parentName:"p",href:"https://github.com/buildbuddy-io/buildbuddy/blob/master/.bazelversion"},".bazelversion")," file."),(0,a.kt)("h2",{id:"docker-image"},"Docker Image"),(0,a.kt)("p",null,"We publish a ",(0,a.kt)("a",{parentName:"p",href:"https://www.docker.com/"},"Docker")," image with every release that contains a pre-configured BuildBuddy."),(0,a.kt)("p",null,"To run it, use the following command:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},"docker pull gcr.io/flame-public/buildbuddy-app-onprem:latest && docker run -p 1985:1985 -p 8080:8080 gcr.io/flame-public/buildbuddy-app-onprem:latest\n")),(0,a.kt)("p",null,"If you'd like to pass a custom configuration file to BuildBuddy running in a Docker image - see the ",(0,a.kt)("a",{parentName:"p",href:"/docs/config"},"configuration docs")," on using Docker's ",(0,a.kt)("a",{parentName:"p",href:"https://docs.docker.com/storage/volumes/"},"-v flag"),"."),(0,a.kt)("p",null,"Note: If you're using BuildBuddy's Docker image locally and a third party gRPC cache, you'll likely need to add the ",(0,a.kt)("inlineCode",{parentName:"p"},"--network=host")," ",(0,a.kt)("a",{parentName:"p",href:"https://docs.docker.com/network/host/"},"flag")," to your ",(0,a.kt)("inlineCode",{parentName:"p"},"docker run")," command in order for BuildBuddy to be able to pull test logs and timing information from the external cache."),(0,a.kt)("h2",{id:"kubernetes"},"Kubernetes"),(0,a.kt)("p",null,'If you run or have access to a Kubernetes cluster, and you have the "kubectl" command configured, we provide a shell script that will deploy BuildBuddy to your cluster, namespaced under the "buildbuddy" namespace.'),(0,a.kt)("p",null,"This script uses ",(0,a.kt)("a",{parentName:"p",href:"https://github.com/buildbuddy-io/buildbuddy/blob/master/deployment/buildbuddy-app.onprem.yaml"},"this deployment file"),", if you want to see the details of what is being configured."),(0,a.kt)("p",null,"To kick of the Kubernetes deploy, use the following command:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},"bash k8s_on_prem.sh\n")),(0,a.kt)("h3",{id:"custom-configuration"},"Custom configuration"),(0,a.kt)("p",null,"Note: the ",(0,a.kt)("inlineCode",{parentName:"p"},"k8s_on_prem.sh")," script requires ",(0,a.kt)("strong",{parentName:"p"},(0,a.kt)("a",{parentName:"strong",href:"https://kubernetes.io/docs/tasks/tools/install-kubectl/"},"kubectl")," version 1.15")," or higher to be installed."),(0,a.kt)("p",null,"To pass in a custom ",(0,a.kt)("a",{parentName:"p",href:"/docs/config"},"config file"),", you can use the ",(0,a.kt)("inlineCode",{parentName:"p"},"-config")," flag:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"bash k8s_on_prem.sh -config my-config.yaml\n")),(0,a.kt)("h3",{id:"output-to-yaml-file"},"Output to yaml file"),(0,a.kt)("p",null,"By default the ",(0,a.kt)("inlineCode",{parentName:"p"},"k8s_on_prem.sh")," script will use ",(0,a.kt)("inlineCode",{parentName:"p"},"kubectl apply")," to deploy BuildBuddy to your current Kubernetes cluster. If you'd like to output the Kubernetes deployment to a yaml file instead that can be checked in, you can use the ",(0,a.kt)("inlineCode",{parentName:"p"},"-out")," flag:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"bash k8s_on_prem.sh -out my-buildbuddy-deployment.yaml\n")),(0,a.kt)("h3",{id:"number-of-replicas"},"Number of replicas"),(0,a.kt)("p",null,"By default the ",(0,a.kt)("inlineCode",{parentName:"p"},"k8s_on_prem.sh")," script will deploy a single replica of BuildBuddy. If you've configured a MySQL database, storage, and other options necessary to support multiple replicas, you can increase the number of BuildBuddy replicas to deploy with the ",(0,a.kt)("inlineCode",{parentName:"p"},"-replicas")," flag."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"bash k8s_on_prem.sh -replicas 3\n")),(0,a.kt)("h3",{id:"restart-behavior"},"Restart behavior"),(0,a.kt)("p",null,"By default the ",(0,a.kt)("inlineCode",{parentName:"p"},"k8s_on_prem.sh")," will restart your BuildBuddy deployment to pick up any changes in your configuration file. This can lead to brief downtime if only one replica is deployed. You can disable this behavior with the ",(0,a.kt)("inlineCode",{parentName:"p"},"-norestart")," flag."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"bash k8s_on_prem.sh -norestart\n")),(0,a.kt)("h3",{id:"enterprise-deployment"},"Enterprise deployment"),(0,a.kt)("p",null,"If you've obtained a BuildBuddy enterprise license, you deploy enterprise BuildBuddy by specifying the ",(0,a.kt)("inlineCode",{parentName:"p"},"-enterprise")," flag."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"bash k8s_on_prem.sh -enterprise\n")),(0,a.kt)("h2",{id:"helm"},"Helm"),(0,a.kt)("p",null,"If you run or have access to a Kubernetes cluster and are comfortable with ",(0,a.kt)("a",{parentName:"p",href:"https://helm.sh/"},"Helm"),", we maintain official BuildBuddy Helm charts that are easy to configure and deploy."),(0,a.kt)("p",null,"They have options to deploy everything necessary to use all of BuildBuddy's bells and whistles - including MySQL, nginx, and more."),(0,a.kt)("p",null,"The official BuildBuddy charts live in our ",(0,a.kt)("a",{parentName:"p",href:"https://github.com/buildbuddy-io/buildbuddy-helm"},"buildbuddy-helm repo")," and can be added to helm with the following command:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"helm repo add buildbuddy https://helm.buildbuddy.io\n")),(0,a.kt)("p",null,"You can the deploy BuildBuddy Open Source with the following command:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"helm install buildbuddy buildbuddy/buildbuddy \\\n  --set mysql.mysqlUser=sampleUser \\\n  --set mysql.mysqlPassword=samplePassword\n")),(0,a.kt)("p",null,"For more information on configuring your BuildBuddy Helm deploy, check out the charts themselves:"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"https://github.com/buildbuddy-io/buildbuddy-helm/tree/master/charts/buildbuddy"},"BuildBuddy Open Source")),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"https://github.com/buildbuddy-io/buildbuddy-helm/tree/master/charts/buildbuddy-enterprise"},"BuildBuddy Enterprise"))),(0,a.kt)("h2",{id:"configuring-buildbuddy"},"Configuring BuildBuddy"),(0,a.kt)("p",null,"For documentation on all BuildBuddy configuration options, check out our ",(0,a.kt)("a",{parentName:"p",href:"/docs/config"},"configuration documentation"),"."))}s.isMDXComponent=!0}}]);