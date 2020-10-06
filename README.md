# Lab1 汉语分词：最大匹配方法

该项目采用Go语言编写，需要在对应的环境执行



## How To Run

1. [安装环境](https://golang.google.cn/doc/install)

2. 设置GOPATH

   1. Linux环境下，输入`export GOPATH='the path that you want to set'`
   2. Windows环境下，设置用户变量GOPATH，重启
   3. Mac环境，没有mac

3. 在GOPATH下放置该项目或者输入`git clone https://gitee.com/shishiredorion/irlab1`

4. 打开项目目录，执行命令`go run main.go` 即可运行

5. 如果喜欢编译，在项目目录下执行`go build` 即可

   

## How To Successfully Run

该项目采用硬编码的方式读取所需的文件，因需要以下三个符合格式要求的文件并放置在工作目录下：

1. corpus.dict.txt

2. corpus.sentence.txt
3. corpus.answer.txt



## How To Check Result

成功执行后，会显示整个流程所需用时，并在工作目录下生成以下两个文件：

1. corpus.ouput.txt，根据词典对语料进行分词后的结果
2. corpus.evaluation.txt，分词的评估结果