# Cosmos Scanning: How We Keep You Safe

Cosmos continuously learns about your perimeter generating an up-to-date map of your attack surface that outpaces the speed and precision of modern adversaries. Cosmos accomplishes this through a combination of scanning and testing activities.

Cosmos scanning is the process of programmatically analyzing your attack surface to uncover potential vulnerabilities that could be exploitable by adversaries.

Large scale scanning has the potential to affect your systems’ performance. Cosmos strikes the balance of scanning more intensively on off-peak hours. We constantly work to innovate to scan more frequently with a reduced system impact.

## Scanning on the updated platform

Here is some information to keep in mind regarding how scanning works in the new platform.

- Scanning is continuous and distributed across every day/week
- We aim to test existence and rescan all existing assets at least once every 24 hours
- We aim to test vulnerability of identified ip and hostname services at least once a week
- New assets are scanned immediately (usually within minutes of adding to the attack surface, sometimes longer)
- Assets that are modified (confidence level, ownership, enabled/disabled) are processed with the same priority as new assets

## Scanning Information

Cosmos testing leverages both automated tools running in the Bishop Fox Platform and manual testing performed by the Cosmos Team. You should expect to see all testing activities coming from the following IP addresses:

Cosmos Operator Test System (Manual Testing)
192.153.76.0/24
64.52.111.0/24

Platform (Automated Testing)
199.201.127.0/24, 64.52.111.248/29

## Platform Automation Behavior

There are two main types of behavior in the platform:

Manual actions - These can alter an asset on the platform, which causes a reaction and generates activity based on the event.

Heartbeat actions - These events ensure all assets are continually reviewed for existence, accuracy, and possible further enumeration.

### Heartbeat Actions

Occur every minute, 5 minutes, 30 minutes, 60 minutes, 12 hours, 24 hours, and every week.

Continuous attack surface scanning (rescan) happens hourly, with a portion of each customer's attack surface scanned based on staleness, that is how long since the last time an asset was scanned or modified.

Continuous rescan runs hourly against approximately 5% of each customers attack surface with a staleness factor of 24 hours.

Continuous rescan also reviews attack surface misses (services that are no longer identifiable) every hour with a staleness of 7 days. (Except DNS which is 48 hours).

Vulnerability scanning is now spread across the entire week instead of only on weekends, to provide better coverage without huge bursts. Vulnerability scanning runs every 5 minutes against approximately 0.1% of the IP and hostname services for each customer. This spreads the scanning out across an entire week.

Please note: Bishop Fox reserves the right to modify, update, or replace any aspect of its scanning schedules at its sole discretion, at any time, and without prior notice to the customers.

## Attack Surface Discovery

|                           |                                                                                                                                        |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **Frequency of Scanning** | **Type of Technology Scanned**                                                                                                         |
| Daily                     | <ul><li>Subdomain Discovery and DNS-resolve</li><li>IP Discovery</li><li>Name Discovery</li><li>Port Probe</li> <li>Whois Lookup </li> |

|

## Cosmos Exposure Identification

## Cosmos Adversarial Testing

|                           |                                       |
| ------------------------- | ------------------------------------- |
| **Frequency of Scanning** | **Type of Technology Scanned**        |
| Monday through Friday     | Manual testing and lead investigation |
| Daily                     | Vulnerability and ad hoc scanning     |

## Scheme/Protocol Coverage

Cosmos’ attack surface discovery capabilities can both confirm reachability and verify scheme/protocol on a given asset for the following services:

|                                                                                                                                                                                                                                                                                                                                       |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| <ul><li>bacnet</li><li>banner</li><li>dnp3</li><li>fox</li><li>ftp</li><li>http</li><li>imap</li><li>ipp</li><li>jarm</li><li>modbus</li><li>mongodb</li><li>mssql</li><li>mysql</li><li>ntp</li><li>oracle</li><li>pop3</li><li>postgres</li><li>redis</li><li>siemens</li><li>smb</li><li>smtp</li><li>ssh</li><li>telnet</li></ul> |

## Port Coverage

Cosmos platform provides port coverage on 2 levels:

There is a list of approximately 50 ports that are actively probed for each known IP address on the attack surface. These are ports that are commonly found to be hosting http based services on our customers’ attack surfaces.

80, 443, 1880, 2082, 2083, 2086, 2087, 4444, 5000, 5001, 5104, 5555, 5601, 5800, 6443, 7000, 7001, 7002, 7999, 8000, 8009, 8080, 8081, 8082, 8089, 8090, 8220, 8443, 8983, 9000, 9080, 9200, 9300, 9443, 9500, 9502, 9508, 9514, 10443, 10250, 15672, 16443, 21240, 21241, 31240, 31241, 50000, 51000, 53211

Additionally, Cosmos platform validates ip:port data it receives daily form 3rd party datasets. Below you will see a list of 1200+ ports that are covered by these 3rd party datasets, these ports are verified for reachability as soon as they enter the platform. Cosmos platform does not actively scan these ports for reachability, it only validates reachability when the data is received from a dataset.

This list may have overlap with our “always checked” internal list.

|       |       |       |       |       |       |       |       |       |       |       |       |       |       |
| ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- |
| 7     | 11    | 13    | 15    | 17    | 19    | 20    | 21    | 22    | 23    | 24    | 25    | 26    | 37    |
| 38    | 43    | 49    | 51    | 53    | 67    | 69    | 70    | 79    | 80    | 81    | 82    | 83    | 84    |
| 85    | 86    | 87    | 88    | 89    | 90    | 91    | 92    | 95    | 96    | 97    | 98    | 99    | 100   |
| 102   | 104   | 106   | 110   | 111   | 113   | 119   | 121   | 123   | 129   | 131   | 135   | 137   | 138   |
| 139   | 143   | 154   | 161   | 175   | 179   | 180   | 195   | 199   | 211   | 221   | 222   | 225   | 263   |
| 264   | 311   | 340   | 389   | 443   | 444   | 445   | 447   | 448   | 449   | 450   | 465   | 491   | 500   |
| 502   | 503   | 515   | 520   | 522   | 523   | 541   | 548   | 554   | 555   | 587   | 593   | 623   | 631   |
| 636   | 646   | 666   | 675   | 685   | 771   | 772   | 777   | 789   | 800   | 801   | 805   | 806   | 808   |
| 830   | 843   | 873   | 880   | 888   | 901   | 902   | 943   | 990   | 992   | 993   | 994   | 995   | 999   |
| 1000  | 1010  | 1012  | 1013  | 1022  | 1023  | 1024  | 1025  | 1026  | 1027  | 1028  | 1029  | 1050  | 1063  |
| 1080  | 1089  | 1099  | 1110  | 1111  | 1119  | 1167  | 1177  | 1194  | 1200  | 1234  | 1250  | 1290  | 1311  |
| 1344  | 1355  | 1366  | 1388  | 1400  | 1433  | 1434  | 1471  | 1494  | 1500  | 1515  | 1521  | 1554  | 1588  |
| 1599  | 1604  | 1650  | 1660  | 1723  | 1741  | 1777  | 1800  | 1820  | 1830  | 1833  | 1880  | 1883  | 1900  |
| 1901  | 1911  | 1935  | 1947  | 1951  | 1962  | 1981  | 1988  | 1989  | 1990  | 1991  | 2000  | 2001  | 2002  |
| 2003  | 2006  | 2008  | 2010  | 2012  | 2018  | 2020  | 2021  | 2022  | 2030  | 2048  | 2049  | 2050  | 2051  |
| 2052  | 2053  | 2054  | 2055  | 2056  | 2057  | 2058  | 2059  | 2060  | 2061  | 2062  | 2063  | 2064  | 2065  |
| 2066  | 2067  | 2068  | 2069  | 2070  | 2077  | 2079  | 2080  | 2081  | 2082  | 2083  | 2086  | 2087  | 2095  |
| 2096  | 2100  | 2111  | 2121  | 2122  | 2123  | 2126  | 2150  | 2152  | 2181  | 2200  | 2201  | 2202  | 2211  |
| 2220  | 2221  | 2222  | 2223  | 2225  | 2232  | 2233  | 2250  | 2252  | 2259  | 2266  | 2320  | 2323  | 2323  |
| 2332  | 2345  | 2351  | 2352  | 2375  | 2376  | 2379  | 2382  | 2404  | 2443  | 2455  | 2480  | 2506  | 2525  |
| 2548  | 2549  | 2550  | 2551  | 2552  | 2553  | 2554  | 2555  | 2556  | 2557  | 2558  | 2559  | 2560  | 2561  |
| 2562  | 2563  | 2566  | 2567  | 2568  | 2569  | 2570  | 2572  | 2598  | 2601  | 2602  | 2626  | 2628  | 2650  |
| 2701  | 2709  | 2727  | 2761  | 2762  | 2806  | 2985  | 3000  | 3001  | 3002  | 3005  | 3048  | 3049  | 3050  |
| 3051  | 3052  | 3053  | 3054  | 3055  | 3056  | 3057  | 3058  | 3059  | 3060  | 3061  | 3062  | 3063  | 3066  |
| 3067  | 3068  | 3069  | 3070  | 3071  | 3072  | 3073  | 3074  | 3075  | 3076  | 3077  | 3078  | 3079  | 3080  |
| 3081  | 3082  | 3083  | 3084  | 3085  | 2086  | 3087  | 3088  | 3089  | 3090  | 3091  | 3092  | 3093  | 3094  |
| 3095  | 3096  | 3097  | 3098  | 3099  | 3100  | 3101  | 3102  | 3103  | 3104  | 3105  | 3106  | 3107  | 3108  |
| 3109  | 3110  | 3111  | 3112  | 3113  | 3114  | 3115  | 3116  | 3117  | 3118  | 3119  | 3120  | 3121  | 3128  |
| 3129  | 3200  | 3211  | 3221  | 3260  | 3270  | 3283  | 3299  | 3306  | 3307  | 3310  | 3311  | 3333  | 3337  |
| 3352  | 3360  | 3386  | 3388  | 3389  | 3391  | 3400  | 3401  | 3402  | 3403  | 3404  | 3405  | 3406  | 3407  |
| 3408  | 3409  | 3410  | 3412  | 3443  | 3460  | 3479  | 3498  | 3503  | 3521  | 3522  | 3523  | 3524  | 3541  |
| 3542  | 3548  | 3549  | 3550  | 3551  | 3552  | 3554  | 3555  | 3556  | 3557  | 3558  | 3559  | 3560  | 3561  |
| 3562  | 3563  | 3566  | 3567  | 3568  | 3569  | 3570  | 3671  | 3689  | 3690  | 3702  | 3749  | 3780  | 3784  |
| 3790  | 3791  | 3792  | 3793  | 3794  | 3838  | 3910  | 3922  | 3950  | 3951  | 3952  | 3953  | 3954  | 4000  |
| 4001  | 4002  | 4010  | 4022  | 4040  | 4042  | 4043  | 4063  | 4064  | 4070  | 4100  | 4117  | 4118  | 4157  |
| 4190  | 4200  | 4222  | 4242  | 4243  | 4282  | 4321  | 4369  | 4430  | 4433  | 4443  | 4444  | 4445  | 4482  |
| 4500  | 4505  | 4506  | 4523  | 4524  | 4545  | 4550  | 4567  | 4588  | 4592  | 4643  | 4646  | 4664  | 4700  |
| 4730  | 4734  | 4747  | 4782  | 4786  | 4800  | 4808  | 4840  | 4848  | 4899  | 4911  | 4949  | 4991  | 4999  |
| 5000  | 5001  | 5002  | 5003  | 5004  | 5005  | 5006  | 5007  | 5008  | 5009  | 5010  | 5025  | 5050  | 5060  |
| 5070  | 5080  | 5090  | 5094  | 5104  | 5122  | 5150  | 5172  | 5190  | 5200  | 5201  | 5209  | 5222  | 5269  |
| 5280  | 5321  | 5353  | 5357  | 5400  | 5431  | 5432  | 5443  | 5446  | 5454  | 5494  | 5500  | 5542  | 5552  |
| 5555  | 5560  | 5567  | 5568  | 5569  | 5577  | 5590  | 5591  | 5592  | 5593  | 5594  | 5595  | 5596  | 5597  |
| 5598  | 5599  | 5600  | 5601  | 5602  | 5603  | 5604  | 5605  | 5606  | 5607  | 5608  | 5609  | 5632  | 5672  |
| 5673  | 5683  | 5684  | 5800  | 5801  | 5822  | 5853  | 5858  | 5900  | 5901  | 5906  | 5907  | 5908  | 5909  |
| 5910  | 5938  | 5984  | 5985  | 5986  | 6000  | 6001  | 6002  | 6003  | 6004  | 6005  | 6006  | 6007  | 6008  |
| 6009  | 6010  | 6036  | 6080  | 6102  | 6129  | 6161  | 6262  | 6264  | 6308  | 6352  | 6363  | 6379  | 6443  |
| 6464  | 6503  | 6510  | 6511  | 6512  | 6543  | 6550  | 6560  | 6561  | 6565  | 6580  | 6581  | 6588  | 6590  |
| 6600  | 6601  | 6602  | 6603  | 6605  | 6622  | 6650  | 6662  | 6664  | 6666  | 6667  | 6668  | 6697  | 6748  |
| 6789  | 6881  | 6887  | 6955  | 6969  | 6998  | 7000  | 7001  | 7002  | 7003  | 7004  | 7005  | 7010  | 7014  |
| 7070  | 7071  | 7080  | 7081  | 7090  | 7170  | 7171  | 7218  | 7401  | 7415  | 7433  | 7443  | 7444  | 7445  |
| 7465  | 7474  | 7493  | 7500  | 7510  | 7535  | 7537  | 7547  | 7548  | 7634  | 7654  | 7657  | 7676  | 7700  |
| 7776  | 7777  | 7778  | 7779  | 7788  | 7800  | 7887  | 7979  | 7998  | 7999  | 8000  | 8001  | 8002  | 8003  |
| 8004  | 8005  | 8006  | 8007  | 8008  | 8009  | 8010  | 8011  | 8012  | 8013  | 8014  | 8015  | 8016  | 8017  |
| 8018  | 8019  | 8020  | 8021  | 8022  | 8023  | 8024  | 8025  | 8026  | 8027  | 8028  | 8029  | 8030  | 8031  |
| 8032  | 8033  | 8034  | 8035  | 8036  | 8037  | 8038  | 8039  | 8040  | 8041  | 8042  | 8043  | 8044  | 8045  |
| 8046  | 8047  | 8048  | 8049  | 8050  | 8051  | 8052  | 8053  | 8054  | 8055  | 8056  | 8057  | 8058  | 8060  |
| 8064  | 8066  | 8069  | 8071  | 8072  | 8080  | 8081  | 8083  | 8084  | 8085  | 8086  | 8087  | 8088  | 8089  |
| 8090  | 8091  | 8092  | 8093  | 8094  | 8095  | 8096  | 8097  | 8098  | 8099  | 8100  | 8101  | 8102  | 8103  |
| 8104  | 8105  | 8106  | 8107  | 8108  | 8109  | 8110  | 8111  | 8112  | 8118  | 8123  | 8126  | 8129  | 8139  |
| 8140  | 8143  | 8159  | 8180  | 8181  | 8182  | 8184  | 8190  | 8200  | 8220  | 8222  | 8236  | 8237  | 8238  |
| 8239  | 8241  | 8243  | 8248  | 8249  | 8251  | 8252  | 8291  | 8333  | 8334  | 8383  | 8401  | 8402  | 8403  |
| 8404  | 8405  | 8406  | 8407  | 8408  | 8409  | 8410  | 8411  | 8412  | 8413  | 8414  | 8415  | 8416  | 8417  |
| 8418  | 8419  | 8420  | 8421  | 8422  | 8423  | 8424  | 8425  | 8426  | 8427  | 8428  | 8429  | 8430  | 8431  |
| 8432  | 8433  | 8442  | 8443  | 8444  | 8445  | 8446  | 8447  | 8448  | 8500  | 8513  | 8545  | 8553  | 8554  |
| 8585  | 8586  | 8590  | 8602  | 8621  | 8622  | 8623  | 8637  | 8649  | 8663  | 8666  | 8688  | 8700  | 8733  |
| 8765  | 8766  | 8767  | 8779  | 8782  | 8784  | 8787  | 8788  | 8789  | 8791  | 8800  | 8801  | 8802  | 8803  |
| 8804  | 8805  | 8806  | 8807  | 8808  | 8809  | 8810  | 8811  | 8812  | 8813  | 8814  | 8815  | 8816  | 8817  |
| 8818  | 8819  | 8820  | 8821  | 8822  | 8823  | 8824  | 8825  | 8826  | 8827  | 8828  | 8829  | 8830  | 8831  |
| 8832  | 8833  | 8834  | 8835  | 8836  | 8837  | 8838  | 8839  | 8840  | 8841  | 8842  | 8843  | 8844  | 8845  |
| 8846  | 8847  | 8848  | 8849  | 8850  | 8851  | 8852  | 8853  | 8854  | 8855  | 8856  | 8857  | 8858  | 8859  |
| 8860  | 8861  | 8862  | 8863  | 8864  | 8865  | 8866  | 8867  | 8868  | 8869  | 8870  | 8871  | 8872  | 8873  |
| 8874  | 8875  | 8876  | 8877  | 8878  | 8879  | 8880  | 8881  | 8883  | 8885  | 8887  | 8888  | 8889  | 8890  |
| 8891  | 8899  | 8935  | 8969  | 8983  | 8988  | 8989  | 8990  | 8991  | 8993  | 8999  | 9000  | 9001  | 9002  |
| 9003  | 9004  | 9005  | 9006  | 9007  | 9008  | 9009  | 9010  | 9011  | 9012  | 9013  | 9014  | 9015  | 9016  |
| 9017  | 9018  | 9019  | 9020  | 9021  | 9022  | 9023  | 9024  | 9025  | 9026  | 9027  | 9028  | 9029  | 9030  |
| 9031  | 9032  | 9033  | 9034  | 9035  | 9036  | 9037  | 9038  | 9039  | 9040  | 9041  | 9042  | 9043  | 9044  |
| 9045  | 9046  | 9047  | 9048  | 9049  | 9050  | 9051  | 9070  | 9080  | 9082  | 9084  | 9088  | 9089  | 9090  |
| 9091  | 9092  | 9093  | 9094  | 9095  | 9096  | 9097  | 9098  | 9099  | 9100  | 9101  | 9102  | 9103  | 9104  |
| 9105  | 9106  | 9107  | 9108  | 9109  | 9110  | 9111  | 9119  | 9136  | 9151  | 9160  | 9189  | 9191  | 9200  |
| 9201  | 9202  | 9203  | 9204  | 9205  | 9206  | 9207  | 9208  | 9209  | 9210  | 9211  | 9212  | 9213  | 9214  |
| 9215  | 9216  | 9217  | 9218  | 9219  | 9220  | 9221  | 9222  | 9251  | 9295  | 9299  | 9300  | 9301  | 9302  |
| 9303  | 9304  | 9305  | 9306  | 9307  | 9308  | 9309  | 9310  | 9311  | 9389  | 9418  | 9433  | 9443  | 9444  |
| 9445  | 9500  | 9502  | 9508  | 9514  | 9527  | 9530  | 9550  | 9595  | 9600  | 9606  | 9633  | 9663  | 9682  |
| 9690  | 9704  | 9743  | 9761  | 9765  | 9861  | 9869  | 9876  | 9898  | 9899  | 9943  | 9944  | 9950  | 9955  |
| 9966  | 9981  | 9988  | 9990  | 9991  | 9992  | 9993  | 9994  | 9997  | 9998  | 9999  | 10000 | 10001 | 10134 |
| 10243 | 10250 | 10443 | 10554 | 11112 | 11211 | 11300 | 12000 | 12345 | 13579 | 14147 | 14265 | 14344 | 15672 |
| 16010 | 16464 | 16992 | 16993 | 17000 | 18081 | 18245 | 18246 | 20000 | 20087 | 20256 | 20547 | 21000 | 21025 |
| 21379 | 22222 | 23023 | 23424 | 25105 | 25565 | 27015 | 27016 | 27017 | 27036 | 28015 | 28017 | 30718 | 32400 |
| 32754 | 32764 | 33060 | 33338 | 37215 | 37777 | 41794 | 44818 | 47001 | 47808 | 48899 | 49152 | 49153 | 49154 |
| 50000 | 50050 | 50070 | 50075 | 50100 | 51000 | 51106 | 51235 | 52869 | 53211 | 53413 | 54138 | 54984 | 55442 |
| 55443 | 55553 | 55554 | 60001 | 60129 | 62078 | 64738 |       |       |       |       |       |       |       |

## Cosmos Scanning Information

### Interactive Services

Cosmos utilizes custom domains for out-of-band interactions on your attack surface. This provides opportunities for attack that may bypass restrictions on raw IP communication. The services that reside on these domains are owned by Bishop Fox, however, they may not resolve back to the manual testing or automated testing ranges mentioned above. If your organization’s staff see interactions with the following domains rest assured the traffic originates from the Cosmos team:

`- *.sploit.in`
`- *.bf.gy`

Self-hosting services allows Bishop Fox to protect client data and prevent third-party traffic monitoring of public servers; customer infrastructure should never be exposed to third -parties using these tools. In addition, self-hosting provides trusted domains to add to an allow list.

### Interactsh Server

Cosmos hosts a private Interactsh server for out-of-band communication while testing various protocols for security vulnerabilities. Payloads from this testing are designed to initiate a callback to servers on the `*.interact.sploit.in` domain. The presence of this domain in server logs is an expected result of testing. This service is passive and does not initiate interactions with your systems. It will only respond to triggered interactions it receives.

### Burp Collaborator Server

Bishop Fox hosts a private Burp collaborator service for testing web applications for security vulnerabilities. Payloads from this testing are designed to initiate a callback to servers on the `*.bf.gy` domain and may not contain any Bishop Fox specific designations in the request. The presence of this domain in server logs is an expected result of testing. This service is passive and does not initiate interactions with your system. It will only respond to triggered interactions it receives.
