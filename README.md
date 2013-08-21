Rubygems Indexer
==========
众说周知现在rubygems数量庞大啊，ruby.taobao.org上大概是三十多万。我们可以通过安装rubygems-mirror这个gem来搭建自己的gem源服务器，安装之后使用`gem mirror`命令就能从上游服务器上面抓取所有的gem包，完了还得生成索引，对应命令是`gem generate_index`，这个命令非常的慢，因为其内部会遍历所有gem，解包并取出其元数据来做索引。在一台8G内存4核8线程的服务器上面跑了10个小时都没有跑完，内存基本耗尽。

既然gem都是从上游服务器抓下来的何不把索引也抓下来呢，于是根据前人的工作（[1]、[2]）加上自己的研究写了这个东西，正好最近在倒腾Golang，就用Golang来做。在同样的机器上面用10个Goroutine并发跑，完成初次索引只用了20多分钟，完成索引更新检查只用了15分钟。你可以修改goroutine的数量来适应你的机器。


	+------------------------------------+
	|       Rubygems Indexer v1.1        |
	|           by horsleyli             |
	+------------------------------------+
	Usage: E:\rubygems_indexer\rubygems_indexer.exe [options]
	
	Options:
	  -d, -destination	the local path which contains index files and the "gems" sub-directory.
	  -s, -source		the upstream rubygems source url.
	  -f				force download all(default update only)


[1]: https://gist.github.com/kcowgill/5526236
[2]: https://github.com/yoshiori/rubygems-mirror-command/blob/master/lib/rubygems/mirror/command/cli.rb