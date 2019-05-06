package ansibleGen

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	yaml "gopkg.in/yaml.v2"
)

//AnsibleProject represents the structure of an Ansible Project
type AnsibleProject struct {
	Name          string
	CustomRoles   []AnsibleRole
	GalaxyRoles   []string
	TreeStructure Folder
	RootDir       string
}

//NewAnsibleProject initializes the structure for a new Ansible project
func NewAnsibleProject(name string, customRoles string, galaxyRoles string) *AnsibleProject {
	ap := &AnsibleProject{
		Name:          name,
		CustomRoles:   splitCustomRoles(customRoles),
		GalaxyRoles:   splitRoles(galaxyRoles),
		TreeStructure: getProjectTreeTemplate(name),
	}
	ap.RootDir, _ = os.Getwd()
	ap.addRolesToTreeStructure()
	ap.addGalaxyRoles()
	return ap
}

//Save run the tree structure creation for the project
func (project *AnsibleProject) Save(dryRun bool) {
	baseFs := afero.NewOsFs()
	fmt.Println("Using root directory: ", project.RootDir)
	WriteTreeToDisk(project.RootDir, project.TreeStructure, &baseFs, dryRun, project.RootDir)
	fmt.Printf("Ansible project %s has been generated\n", project.Name)
}

//InitGit initializes a Git repository
func (project *AnsibleProject) InitGit(dryRun bool) {
	if !dryRun {
		cmd := "git"
		args := []string{"init", filepath.Join(project.RootDir, project.Name)}
		if err := exec.Command(cmd, args...).Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	fmt.Println("Successfully initialized a Git repository for the project")
}

func (project *AnsibleProject) addGalaxyRoles() {
	if len(project.GalaxyRoles) > 0 {
		var galaxyRolesFile = File{Name: "galaxy-roles.yml"}
		fileContent := "---\n"
		for _, role := range project.GalaxyRoles {
			fileContent += fmt.Sprintln("-", role)
		}
		galaxyRolesFile.Content = fileContent
		project.TreeStructure.Files = append(project.TreeStructure.Files, galaxyRolesFile)
	}
}

func (project *AnsibleProject) addRolesToTreeStructure() {
	rolesIndex := project.rolesFolderIndex("roles")
	rolesFolder := &project.TreeStructure.Folders[rolesIndex]
	for _, role := range project.CustomRoles {
		rolesFolder.Folders = append(rolesFolder.Folders, role.TreeStructure)
	}
}

func (project *AnsibleProject) rolesFolderIndex(rolesFolderName string) int {
	for index, folder := range project.TreeStructure.Folders {
		if folder.Name == rolesFolderName {
			return index
		}
	}
	return 0
}

func splitCustomRoles(customRoles string) []AnsibleRole {
	roles := []AnsibleRole{}
	for _, roleName := range splitRoles(customRoles) {
		roles = append(roles, *NewAnsibleRole(roleName))
	}
	return roles
}

func splitRoles(roles string) []string {
	if len(roles) == 0 {
		return nil
	}
	return strings.Split(roles, ",")
}

func getProjectTreeTemplate(projectName string) Folder {
	projectTemplate := `
---
name: ` + projectName + `
files:
- name: ansible.cfg
folders:
- name: playbooks
- name: inventories
  folders:
  - name: dev
    folders:
    - name: group_vars
      files:
      - name: group1.yml
      - name: group2.yml
    - name: host_vars
      files:
      - name: host1.yml
      - name: host2.yml
  - name: production
    folders:
    - name: group_vars
      files:
      - name: group1.yml
      - name: group2.yml
    - name: host_vars
      files:
      - name: host1.yml
      - name: host2.yml
  - name: staging
    folders:
    - name: group_vars
      files:
      - name: group1.yml
      - name: group2.yml
    - name: host_vars
      files:
      - name: host1.yml
      - name: host2.yml
- name: roles
`

	dst := Folder{}
	err := yaml.Unmarshal([]byte(projectTemplate), &dst)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return dst
}
