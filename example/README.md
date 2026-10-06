# ASUN vs JSON

**all samples from https://sample.json-format.com/**

| ASUN(Byte) | JSON (Byte) | Save   | Name                                                                |
| ---------- | ----------- | ------ | ------------------------------------------------------------------- |
| 4933       | 10230       | 51.78% | 1MB_Crypto Historical 365 Days { 1 } Level Minified Versions        |
| 3128773    | 10493514    | 70.18% | 10MB_Cumulative { 1 } Level Minified Versions                       |
| 3471497    | 10492854    | 66.92% | 10MB_EAFC26 Men { 1 } Level Minified Versions                       |
| 1914007    | 10500506    | 81.77% | 10MB_Ecommerce Customer Churn Dataset { 1 } Level Minified Versions |
| 8647307    | 10486655    | 17.54% | 10MB_Employees { 5 } Level Nested Formatted Versions                |
| 8804655    | 10562056    | 16.64% | 10MB_Final Dataset { 1 } Level Minified Versions                    |
| 6312557    | 10502015    | 39.89% | 10MB_Games { 1 } Level Minified Versions                            |
| 4402104    | 10506983    | 58.1%  | 10MB_Global Earth Quakes 10-Years { 1 } Level Minified Versions     |
| 5623342    | 10511483    | 46.5%  | 10MB_Listening History { 1 } Level Minified Versions                |

Sizes are of the files in this folder. The Employees files are indented, which
costs ASUN relatively more than JSON; minified, the same data is 2044911 B of
ASUN against 4144323 B of JSON (50.66% saved).
