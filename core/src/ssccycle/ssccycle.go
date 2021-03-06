package ssccycle

import (
	"txffc/core/model"
	"fmt"
	"strings"
	"time"
	"log"
	"txffc/core/logger"
	"strconv"
	"txffc/core/mail"
)

// 数据包
var contain_datapackage []*model.SscCycle

// 最新开奖号
var new_code string = ""

type multipleData struct {
	packageA  	map[string]string
	code 		[]string
	position   	string
	packet      *model.SscCycle
}

// 计算
// 时时彩 a包连续 周期计算
func Calculation()  {
	// 查询
	txffc := new(model.Txffc)

	// 最新开奖号
	newcode := txffc.GetNewsCode()

	// 是否重复计算
	if new_code != "" && new_code == newcode {
		fmt.Println("最新一期 腾讯分分彩 已经计算过了 等待新号码出现")
		return
	}

	// 刷新 最新开奖号
	new_code = newcode

	// 查询数据包
	sscdata := new(model.SscCycle)
	contain_datapackage = sscdata.Query()

	for i := range contain_datapackage {
		go containAnalysisCodes(contain_datapackage[i])
	}

}

func containAnalysisCodes(packet *model.SscCycle)  {
	slice_dataTxt := strings.Split(packet.DataTxt, "\r\n")
	//slice data txt to slice data txt map
	dataTxtMap := make(map[string]string)
	for i := range slice_dataTxt {
		dataTxtMap[slice_dataTxt[i]] = slice_dataTxt[i]
	}

	//检查是否在报警时间段以内
	if (packet.Start >0 && packet.End >0) && (time.Now().Hour() < packet.Start || time.Now().Hour() > packet.End)  {
		log.Println("彩票类型: 腾讯分分彩 a包连续周期", "数据包别名:", packet.Alias, "报警通知非接受时间段内")
		logger.Log("彩票类型: 腾讯分分彩 a包连续周期 数据包别名: " + packet.Alias + " 报警通知非接受时间段内 ")
		return
	}

	// 查询 200期内开奖号
	txffcModel := new(model.Txffc)
	codes := txffcModel.Query("400")

	q3codes := make([]string, 0)
	z3codes := make([]string, 0)
	h3codes := make([]string, 0)
	for i := range codes {
		q3s := codes[i].One + codes[i].Two + codes[i].Three
		z3s := codes[i].Two + codes[i].Three + codes[i].Four
		h3s := codes[i].Three + codes[i].Four + codes[i].Five
		q3codes = append(q3codes, q3s)
		z3codes = append(z3codes, z3s)
		h3codes = append(h3codes, h3s)
	}


	q3 := &multipleData{
		packageA: dataTxtMap,
		code: q3codes,
		position: "前三",
		packet: packet,
	}

	z3 := &multipleData{
		packageA: dataTxtMap,
		code: z3codes,
		position: "中三",
		packet: packet,
	}

	h3 := &multipleData{
		packageA: dataTxtMap,
		code: h3codes,
		position: "后三",
		packet: packet,
	}

	go q3.calculate()
	go z3.calculate()
	go h3.calculate()
}

func (md *multipleData) calculate() {
	//md.code = make([]string, 0)
	//md.code = append(md.code, "661")
	//md.code = append(md.code, "577")
	//md.code = append(md.code, "065")
	//md.code = append(md.code, "345")
	//md.code = append(md.code, "411")
	//md.code = append(md.code, "020")

	// 周期数
	var cycle_number int = 0

	// a包连续数
	var continuity_number int = 0

	// a连续完 b在规定期数
	var b_number int = 0

	// 重新累计 连续a
	var a_status bool = true

	// 1周期累计完了后 直到b出现后才累计下一周期
	var next_cycle_status bool = false

	// 最新的一期是否是A包
	var in_a_status bool = false

	var log_html string = "<div>腾讯分分彩 a连续周期 包别名: " + md.packet.Alias + " 位置: "+ md.position + "<div>"
	for i := range md.code {
		log_html += "<br/><div>开奖号: " + md.code[i] + "</div>"

		// 检查是否包含a包
		_, in_a := md.packageA[md.code[i]]

		//检查上一期是否 包含A包
		pre_code := ""
		if i != 0 {
			pre_code = md.code[i - 1]
		}

		var pre_in_a bool = false
		//上一期是否包含A包
		if pre_code != "" {
			_, pre_in_a = md.packageA[pre_code]
		}

		log_html += "<div>本期包含a包:"+ strconv.FormatBool(in_a) +"</div>"
		log_html += "<div>上期包含a包:"+ strconv.FormatBool(pre_in_a) +"</div>"

		if in_a {
			in_a_status = true
		} else {
			in_a_status = false
		}

		// 1周期累计完了后 直到b出现后才累计下一周期
		if next_cycle_status == false && in_a == false {
			next_cycle_status = true
			log_html += "<div> 一个整周期累计完 直到出现b后才开始累计下一周  周期后 b出现 开始下一周期累计 </div>"
			continue
		}

		// 1周期累计完了后 直到b出现后才累计下一周期
		if next_cycle_status == false && in_a == true {
			log_html += "<div> 一个整周期累计完 直到出现b后才开始累计下一周  周期后 未出现b 等待b出现 </div>"
			continue
		}

		// 第一期 出现a包 算 1连续
		if i == 0 && in_a == true {
			continuity_number += 1
			log_html += "<div>第一期 就 包含a包 连续数+1 = " + strconv.Itoa(continuity_number) + "</div>"
			continue
		}

		// a包连续到达 阀值
		if continuity_number >= md.packet.Continuity {
			b_number += 1
			log_html += "<div> 等待b包在规定期数内出现 b开始累计 = " + strconv.Itoa(b_number) + "</div>"

			// a包连续完 在b包规定期数内出现了b
			if b_number <= md.packet.Bnumber && in_a == false {
				// 周期清零
				cycle_number = 0
				// 连续清零
				continuity_number = 0
				// b连续期数清零
				b_number = 0
				// 重新计算连续a
				a_status = true
				log_html += "<div> a包连续完 在b包规定期数内出现了b </div>"
				log_html += "<div> 周期值: "+ strconv.Itoa(cycle_number) +" </div>"
				log_html += "<div> a连续值: "+ strconv.Itoa(continuity_number) +" </div>"
				log_html += "<div> b值: "+ strconv.Itoa(b_number) +" </div>"
			}

			// a包连续完 未在b包规定期数内出现了b
			if b_number > md.packet.Bnumber {
				// 周期+1
				cycle_number += 1
				// 连续清零
				continuity_number = 0
				// b连续期数清零
				b_number = 0
				// 重新计算连续a
				a_status = true

				// 1周期累计完了后 直到b出现后才累计下一周期
				next_cycle_status = false

				log_html += "<div> a包连续完 未在b包规定期数内出现了b 周期+1 = "+ strconv.Itoa(cycle_number) +" </div>"
			}
			continue
		}

		// 包含a包 且 上一期 包含a包
		if in_a == true && pre_in_a {
			continuity_number += 1
			// 关闭 重新计算连续a
			a_status = false
			log_html += "<div>本期 包含a包 且 上期包含a包 连续数+1 = " + strconv.Itoa(continuity_number) + "</div>"
			continue
		}

		// 本期b包
		if in_a == false {
			// 连续清零
			continuity_number = 0
			// b连续期数清零
			b_number = 0
			// 重新计算连续a
			a_status = true

			log_html += "<div> 本期b包 a连续值 " + strconv.Itoa(continuity_number) + " </div>"
			continue
		}

		// 重新累计连续a
		if i > 0 && in_a == true && a_status == true {
			continuity_number += 1
			log_html += "<div> 重新累计连续a包 " + strconv.Itoa(continuity_number) + " </div>"
			continue
		}
	}

	// 检查是否报警
	if cycle_number >= md.packet.Cycle - 1 && continuity_number == md.packet.Continuity && in_a_status {
		body_html := "<div>腾讯分分彩 a连续b周期 报警 位置: "+ md.position+ " 数据包别名: "+ md.packet.Alias+ " 几A几B: " + strconv.Itoa(md.packet.Continuity) + " A " + strconv.Itoa(md.packet.Bnumber) + " B " + " 当前累计周期数 "+ strconv.Itoa(cycle_number) + " 当前a连续: "+ strconv.Itoa(continuity_number) +"</div>"
		body_html += log_html
		go mail.SendMail("腾讯分分彩 a连续b周期 报警", body_html)
	}
}