# Kubernetes在中国的生产量占72％

在CNCF，我们定期调查社区，以更好地了解开源和云原生技术的采用。我们第三次使用中文进行了"云原生调查中国"，以更深入地了解中国采用云原生的速度，以及如何在这个庞大且不断发展的社区中增强开发人员的能力并改变其发展。本报告以2018年3月和2018年11月发布的前两份中国报告为基础。

# 中国云原生调查的重点
49％的受访者在生产中使用容器，另有32％的人计划这样做。与2018年11月相比，这是一个显着的增长，当时生产中仅使用了20％的容器。
72％的受访者在生产中使用Kubernetes，高于2018年11月的40％。 
公共云的使用率从2018年11月的51％下降到36％，取而代之的是使用39％的混合新选项。
CNCF项目的使用呈指数增长。CNCF现在主持了四个在中国诞生并在该地区更广泛使用的项目：正在孵化的Dragonfly和KubeEdge，以及刚毕业的Harbor和TiKV。 

《 2019年中国云原生调查》共收到300人的回应-其中97％来自亚洲，主要是中国。 

# 容器使用
我们知道容器已经改变了基于云的基础架构，但是在过去的一年中，容器在生产中的使用已成为常态。根据我们今年早些时候发布的全球2019年Cloud Native调查，有84％的受访者在生产中使用容器，这使得容器在全球范围内无处不在。 

对中国的调查表明，尽管中国的容器使用量落后于全球采用率，但其势头正在增强。在中国调查中，将近一半(49％)的受访者在生产中使用了容器–从我们2018年3月的调查中的32％和2018年11月的20％跃升至更高。 

计划在生产中使用容器的中国会员少得多-我们在2018年3月的调查中为57％，在11月为40％。这意味着许多组织已将容器计划付诸实施，而不再处于计划阶段。但是，仍然存在增长的空间，我们完全希望继续增长。

![](http://img.rocdu.top/20201020/BBSZ4HEawaQmk_fMS7UHD6X2tlckq33nqZbOI8qssGPYtZcCGxw-eiRCPVFyA9JLKGDXGHW9ycy5Sre--XLdEqvQNEpJiUoYNpEDJ_6uTN11Q3x0XPJ21n4CKXfctQL-T4_GXmuT)

随着生产用途的增加，测试环境中容器的存在已减少。约28％的中国调查受访者目前正在使用容器进行测试-与2018年3月的24％相比略有上升，但与2018年11月的调查中的42％相比有所下降。

尽管容器带来了惊人的优势，但它们仍然带来了挑战。随着时间的推移，这些已经发生了变化，但是复杂性的挑战一直保持不变。在中国调查中，有53％的受访者将复杂性称为最大挑战-相比之下，我们在2018年3月的调查中将44％的调查列为最高挑战，在2018年11月的调查中将28％的调查列为最高挑战，排名第三。

在挑战方面，安全性排名第二，占39％。这是安全首次被列为首要挑战。缺少培训和网络的比例为36％，排在第三位，而35％的调查受访者则选择可靠性和监视作为部署挑战。


![](http://img.rocdu.top/20201020/Ok8Gv56_Tb3z7Al05olnjjUVOAhLoXi0oJFvxh37BEIF6V4VTzkGkMRXlLYONXF17Nz6YYs2e9q3w-1NJ1kZHf7fg-dfc4yEJA7p97TJ6Y9gg0l_KCQlGvttPENuNM0QYo5uKwcB)
 
# Kubernetes增长

Kubernetes作为一个容器编排的通用平台正在行业中崭露头角，而中国的CNCF社区的采用率也急剧上升。72％的受访者表示在生产中使用Kubernetes-与2018年11月的40％相比大幅增长

因此，评估Kubernetes的人数下降了，从42％降至17％。

![](http://img.rocdu.top/20201020/kT_4KCkv938sXjQPfeVhaY_mms3YRFUxCMqrdleCF8PYKmM3BFrwvpZCr5y5ijtrUR81Y1mRqYfd4hIi3OpJFpNG-VHbH7uMkQJZsB_Ol4YOiagSJbTYr5KvtVF9JW5QyZGGiaJo)

我们还看到Kubernetes的生产集群在部署范围的两端都在增长。大部分回应中国调查的组织使用的集群少于10个，但是运行在50个以上的集群的数量有所增加。这可能是由于在生产中使用容器的新受访者数量增加，而使用生产容器的受访者数量增加了集群。 

36％的受访者拥有2到5个集群，高于2018年11月的25％，一半的人使用1到5个集群，70％的在1到10之间。只有13％的受访者拥有超过50个集群，而2018年11月 5％的人拥有5％的集群。 

![](http://img.rocdu.top/20201020/aMuj-V9UySCSMRXPxrZv7ANL70JfBgLoq3J8nh-n2IbUWlMYTXQs0FW96BC53FsYuCYefOv--f8lKK8gpRxdKpiQEuDl_k28UslfGpQzwfqR930w_Lj4cr4evap1iZC7eITbZMaQ)

# 打包

Helm是包装Kubernetes应用程序的最受欢迎的方法，有54％的受访者选择了这种方法。

# Ingress
NGINX(54％)是使用最多的Kubernetes入口提供商，其次是HAProxy(18％)，F5(16％)和Envoy(15％)。 


# 分离Kubernetes应用程序 
在集群中管理对象可能是一个挑战，但是名称空间通过将它们作为组进行过滤和控制来提供帮助。71％的受访者将其Kubernetes应用程序与名称空间分开。在多个团队中使用Kubernetes的公司中，有68％使用名称空间。

# Monitoring,Logging和Tracing

对于那些使用监视，日志记录和跟踪解决方案的用户来说，它们是在本地运行还是通过远程服务器托管。46％的受访者使用本地监控工具，而20％的受访者通过远程服务运行它们。整体上使用日志记录和跟踪的受访者较少，但是26％的受访者在本地运行跟踪，而通过远程服务运行的是20％。21％的企业内部运行跟踪工具，另外21％的企业通过远程服务运行。

# 代码

得益于持续集成(CI)和持续交付(CD)的支持，云和容器的强大功能共同推动了中国的开发和部署速度。我们的调查通过开发人员将代码检入存储库的频率来量化开发速度。每天有35％以上的代码每天要多次检入代码。每周几次几次检入43％的代码，每月一次几次检入16％的代码。 

![](http://img.rocdu.top/20201020/gEOHlKXd-Dz0PZ2BGED4EOlsN3Gd92_aS4nZ_JaJ2e4RLVc7K7Lp42o6I6q-P9DarDP6vs-3ieUmFeSVzRhWfz3dHzuKaFNshmlg1T6V8YxKys4nfdWDs7dLsCHXwusT96gR_vaJ)

大多数受访者以每周一次的发布周期(43％)工作，而仅五分之一(21％)的工作以每月周期进行，而18％的则以每日周期进行。12％的被调查者按临时时间表工作。

![](http://img.rocdu.top/20201020/HOwkHZ3PGr46ZF3u5opVwS3oGJLpdNFuEsK5NuDglNlUcJtxCumYoCWeYsy3FOUktk6h4TkReZqr9nsB9bNR6gmJ-lWaioeBwhcoMapeyrc99ynCbjHtVfsN91W29-iu_r38NFQL)

# CI/CD

许多人认为成功的CI/CD的基础是流程的自动化。但是，我们在中国的调查显示，纯自动化环境相对较少-只有21％的企业采用自动发布周期，而31％的企业则依靠手动流程。最受欢迎的回应是混合模式，占46％。
![](http://img.rocdu.top/20201020/NHzjrvyXjFhHqNqwncqh3q0pmk3suXih5oPuGaKGjE-M1r9y9aX8lA1sG3cY2XEOZf9ySkwwq6M2XP1tJDLi3wRhP66hBE4K3wqzSh86AJuhyTe7Q4wHnbK0xlTjPUxDlpzMxPh9)

CI/CD是一种哲学和技术，可实现云原生系统的灵活，灵活的交付和生命周期管理。Jenkins是中国社区中最受欢迎的CI/CD工具，仅占社区的一半以上，占53％，而GitLab则占40％。

# 云与内部部署

云在增长，但是今年的中国调查显示，云已经从公共云，私有云的合并以及混合云的出现转变了。在我们的2018年11月调查中，公共云的使用似乎已达到峰值，达到51％，而今年下降到36％。私有云保持稳定，从2018年11月的43％增长到42％。混合云是今年的新选择，占39％。 
![](http://img.rocdu.top/20201020/5j9hbIniirgnZRpkgM-ogEcemyWbIkz4XYsFCm73R_A-QRcAOtywyYYoYaRpWFgLhJwWoCuEfGO_AImT75Tb_kQ-ipM7Oru5NiFkDjYx5QgPhntS-BgbMD4bvxM8CeL7JscDhfvK)

# 云原生项目
CNCF管理着大量的开源项目，这些项目对于云本机的开发，部署和生命周期管理至关重要。CNCF项目在中国呈指数级增长。例如，有57％的受访者使用了Prometheus监视和警报系统，比2018年3月的16％显着增加。现在，CoreDNS的使用率为35％，高于2018年3月的10％。容器化运行时也实现了惊人的增长–从2018年3月的3％增加到2019年初的29％。 

CNCF还主持了在中国创建的四个项目，这些项目在该地区得到了更广泛的应用。蜻蜓(生产中使用量占17％)和KubeEdge(生产中使用量占11％)是两个使用最多的沙盒项目。现在两个都在孵化项目。Harbor和TiKV是毕业项目，分别用于生产的占27％和5％。 

![](http://img.rocdu.top/20201020/u4RISenFfPgtQQKyz7-xNYHcqpJnzQm3E5j0RRT1JX1ynHnq4tXzZJY_VYNu0DmfCm93YwI5ycanBvEFvtH_rhOp_pp65FwS5ATIQRsHuNPiMLDMnoslUnv0v7bCwR1PHWyvC5PD)

>不包括：最近毕业的Rook

![](http://img.rocdu.top/20201020/TfUfEmCoU0YmyMtqregbs6B2GfIG3LkN_GLXtTDSuiLeir-E4GgXVVMiMLRzflspDxZwYAgHgWic2WFZ5hX5bNBufUjRdfEfg6bgT6nR0pRFw8uE28F5rgkkc90e_3do5uxbCDpU)

> 不包括：新的孵化项目Argo，Contour和Operator框架。Rook现在是一个毕业项目。

自CNCF上次在中国进行调查以来，在生产中使用云原生项目带来的好处已发生了转变： 

47％的受访者认为，更快的部署时间首次成为最大的收益。 
改进的可扩展性保持了其早期的第二位，为35％。 
成本节省仍然排名第三，为33％。 
提高的开发人员生产力，云可移植性和更高的可用性以31％排名第四。在2018年11月，可用性已排名第一，可移植性排名第四。 

# 无服务器

![](http://img.rocdu.top/20201020/8UZWBUSSORgOGocrhJjT39DJLYqLXdTMBm9SzsKhV1_KbzNdNy8tqXj3IG7X7R8ueuQGfP4Sbm0YSySPPjVZO1_rJQkzdyqfOi3zzFk3Z3obcNelWQ_1HXyq3lDVndRJ39pYiB9C)

在中国的调查中，无服务器被36％用作托管平台，被22％的用户用作可安装软件。

![](http://img.rocdu.top/20201020/USUup9E7ZlnNL5vOodFbVh1tuY2W1rT7mG3wJ6GVgqRkql8NvYv9LtKQ9lB9sXaIkVzUhDqD47_XE_FvZ0Ad93T_1ICG6ohrBvOG1sbwqpzp6jP-XHoxOl3YiBHpjQ9xsAR2JY1J)

对于那些使用无服务器工具作为托管平台的企业，排名前三的提供商是阿里云功能计算(46％)，AWS Lambda(34％)，以及腾讯云无服务器云功能和华为FunctionStage之间的并列关系(12％)。

![](http://img.rocdu.top/20201020/5ro6rjDtx1vDdwEgvjp4q2sjmGlUkSyqEMJiis9glGnrbv6OMtwcLp-_l_ZEJBxJRTqeai7XXHejrSSOpnmpH1zTvkvfvIpmYaDPVkM4hucBPXtej56wez6cwdvODCcwd7dZ8Xem)

 对于那些使用无服务器工具作为可安装软件的用户，Kubeless排名第一(29％)，其次是Knative(22％)，以及Apache OpenWhisk(20％)。
 
 对于2019年，我们在云原生存储和服务网格上添加了新问题。以下是流行的云原生项目，这些项目在活跃的生产环境中巩固了这些优势：
 
# 存储

![](http://img.rocdu.top/20201020/ov8BiQi9_tdBTs_Wh_TqScVYS0cQy-IAuyLuupxlhHjGlK3Afu_B1OpW6FO9-myQiQOwtImQM5wnvQNzE74C8DD5w2JedTYe80MKqilRIp4ut2l89bBQikJtF7dGKXxcSIKFjzet)

最常用的云原生存储项目是Ceph(24％)，Amazon弹性块存储(EBS)(23％)和容器存储接口(CSI)(18％)

# 服务网格

![](http://img.rocdu.top/20201020/gql8EkWGGMBPUoe7dgWZxTtMGkbXZNQXvLYWzMfccxssEYGebkwur3N8DLqJ9Ui7Zp6HvbH_rmJTtwsGRZZKP9D6Qzxz46SACtsWMeTug8VrIwDJYQdjy6xs6TSImMSQa3ZnoQtY)

# 中国云原生社区

CNCF现在在中国有近50个成员。中国还是CNCF项目的第三大贡献者(按贡献者和提交者计)，仅次于美国和德国。

我们有一些来自中国公司的案例研究，包括：

- 京东(JD.com)为其使用Harbor的私有映像中央存储库节省了大约60％的维护时间。
- 中国民生银行，其交付效率提高了3-4倍，并且使用Kubernetes的资源利用率翻了一番。 
- 蚂蚁金服(Ant Financial)，在使用云原生技术的运营方面至少获得了十倍的改进。

我们还在中国开设了20,000多人参加的Kubernetes和Cloud Native课程，最近还完成了首届Cloud Native +开源虚拟峰会中国。 

中国社区以多种不同方式了解云原生技术。

![](http://img.rocdu.top/20201020/Y23tlzuGt_MgLuthmOhQTK-eWE-ukhi_Rs7DgPTXZ-ozDRPJpWbYaZNCVCVZLfqkdKaqzg9I9lpQMC13BNbS9Ijh4OmNVyWdMXYat2oQP56XSiOn32VcF0zHyMSQbJeK60T_xZKE)

# 文献资料

72％的中国受访者通过文档了解了云原生技术。每个CNCF项目在其网站上都有大量文档，可在此处找到。 

CNCF每年投资数千美元来改善项目文档。这包括托管项目文档，添加教程，操作指南等。

# 大事记 

活动是受访者了解云原生技术的一种流行方式。 

41％的受访者选择KubeCon + CloudNativeCon作为学习新技术的地方。下一个虚拟KubeCon + CloudNativeCon计划于11月17日至20日举行。

37％的受访者选择聚会和本地活动(例如Cloud Native Community Groups)作为了解云原生技术的方式。

# 网络研讨会

22％的受访者通过技术网络研讨会了解云原生技术，另有8％的企业选择面向业务的网络研讨会，还有8％的消费者选择CNCF网络研讨会。

CNCF加强了其网络研讨会计划，并计划为中国观众安排定期的网络研讨会。您可以在此处查看即将到来的日程安排，并查看以前的网络研讨会的录像，幻灯片和重放。 

## 关于调查方法和受访者

非常感谢参与此调查的每个人！

该调查于2019年10月进行。该调查以普通话进行，在300名受访者中，有97％来自亚洲。


![](http://img.rocdu.top/20201020/VtLNtsWfW4ujWwUAEtFmgKlQ5yJ1jFxwtg5XZKhvCXwJe7lT7DKlgBlIIJV1KUHxSyB7VqjDd4YN-gbUdkfg1cgZ7jCzAQjuMDPxLFiBMlCd41ontYtgAVhyjP7SonMTfjXMTQ0k)

![](http://img.rocdu.top/20201020/charts-cncf-china-survey_question3-1024x731.png)

![](http://img.rocdu.top/20201020/ItDX7Y3KgRTh-4-Qa_-YQNcJJTr-eKghFPhezkZvldy9pvTKL70eel72xEqCP-7kfMSIcZl416JIMbMgHaFBoJRZabBVsQytZ2_0hNN3iLgVanyM77edtUh5QGodmqJEYYRntYn9)

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
